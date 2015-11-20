package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
)

var stmtSnapUser *sql.Stmt
var stmtTopAnaswer *sql.Stmt

func snapUser() {
	var err error
	stmtSnapUser, err = db.Prepare(`INSERT INTO usersnapshots(sid, id, followee, follower, ask, answer, post, fav, edits, agree, thanks, date)
									VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		panic(err)
	}
	defer stmtSnapUser.Close()

	stmtTopAnaswer, err = db.Prepare(`INSERT INTO usertopanswers(sid, id, title, agree, date, answerid, link, ispost, noshare, len, summary)
									VALUES(?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		panic(err)
	}
	defer stmtTopAnaswer.Close()

	var sid, finished int
	err = db.QueryRow("select sid, finished from snapshot_config order by tid desc limit 1").Scan(&sid, &finished)
	if err != nil {
		panic(err)
	}

	log.Printf("sid:%d, finished:%d", sid, finished)

	if sid == 0 || finished == 1 {
		sid++
		if result, err := db.Exec("insert into snapshot_config(sid, startAt) values(?,?)", sid, time.Now()); err == nil {
			if id, err := result.RowsAffected(); err == nil {
				log.Println("RowsAffected : ", id)
			}
		} else {
			log.Fatal("err:", err)
		}
	}

	var count int
	err = db.QueryRow("select count(id) from users where id not in (select id from usersnapshots where sid=?) order by tid", sid).Scan(&count)
	if err != nil {
		panic(err)
	}

	rows, err := db.Query("select id from users where id not in (select id from usersnapshots where sid=?) order by tid", sid)
	if err != nil {
		panic(err)
	}

	log.Printf("%d users to be proccessed", count)

	routine := 15
	var gpool GoroutinePool
	gpool.Init(routine, count)

	for rows.Next() {
		var id string

		rows.Scan(&id)

		log.Println(sid, id)

		gpool.AddTask(func() error {
			return doSnapshotSingleUser(sid, id)
		})
	}
	rows.Close()

	now := time.Now()
	gpool.Start()

	finished = 1
	if result, err := db.Exec("UPDATE snapshot_config set endAt=?, finished=? WHERE sid=?",
		time.Now(), finished, sid); err == nil {
		if id, err := result.RowsAffected(); err == nil {
			log.Println("SetFinishCallback RowsAffected : ", id, sid, finished, time.Now())
		}
	} else {
		log.Println("err:", err)
	}

	runTime := time.Now().Sub(now)
	log.Println("usersnapshots task finished in ", runTime)

	doSavePage()

	gpool.Stop()
}
func doSnapshotSingleUser(sid int, id string) error {
	resp := doPageRequest("http://www.zhihu.com/people/" + id)
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return err
	}

	sidebar := doc.Find(".zu-main-sidebar .item")
	followee := sidebar.Eq(0).Find("strong").Text()
	follower := sidebar.Eq(1).Find("strong").Text()

	navbar := doc.Find(".profile-navbar .num")
	ask := navbar.Eq(0).Text()
	answer := navbar.Eq(1).Text()
	post := navbar.Eq(2).Text()
	fav := navbar.Eq(3).Text()
	edits := navbar.Eq(4).Text()

	agree := doc.Find(".zm-profile-header-user-agree").Find("strong").Text()
	thanks := doc.Find(".zm-profile-header-user-thanks").Find("strong").Text()

	name := doc.Find(".zm-profile-header-main .title-section .name").Text()
	avatar, _ := doc.Find(".zm-profile-header-avatar-container .avatar").Attr("src")
	description := doc.Find(".zm-profile-header-description .fold-item .content").Text()

	var sex interface{}
	class, _ := doc.Find(".info-wrap .item .icon").Attr("class")

	g := strings.Split(class, "-")
	gender := g[len(g)-1]
	if gender == "female" {
		sex = 0
	} else if gender == "male" {
		sex = 1
	} else {
		sex = nil
	}

	if name == "" || avatar == "" {
		return errors.New("no name or avatar found, retry")
	}

	h := md5.New()
	h.Write([]byte(id))
	h.Write([]byte(name))
	h.Write([]byte(gender))
	h.Write([]byte(description))
	hash := fmt.Sprintf("%s", hex.EncodeToString(h.Sum(nil)))

	if result, err := db.Exec("UPDATE users set hash=?, name=?, avatar=?, sex=?, description=? WHERE id=? and hash !=?",
		hash, name, avatar, sex, description, id, hash); err == nil {
		if id, err := result.RowsAffected(); err == nil {
			log.Println("RowsAffected : ", id)
		}
	} else {
		log.Println("err:", err)
		return err
	}

	log.Printf("hash:%s, id:%s, gender:%s, sex=%d, avatar:%s", hash, id, gender, sex, avatar)

	now := time.Now()
	log.Printf("sid:%d, followee:%s, follower:%s, ask:%s, answer:%s, post:%s, fav:%s, edits:%s, agree:%s, thanks:%s, date:%s",
		sid, followee, follower, ask, answer, post, fav, edits, agree, thanks, now)

	if result, err := stmtSnapUser.Exec(sid, id, followee, follower, ask, answer, post, fav, edits, agree, thanks, now); err == nil {
		if id, err := result.LastInsertId(); err == nil {
			log.Println("insert id : ", id)
		}
	} else {
		log.Println("err:", err)
		return err
	}

	answerNum, _ := strconv.Atoi(answer)
	return doSnapAnswer(sid, id, answerNum)
}

func doSnapAnswer(sid int, id string, answer int) error {
	var err error
	page := answer/20 + 1

	var shouldBreak bool
	for i := 1; i <= page; i++ {
		url := fmt.Sprintf("http://www.zhihu.com/people/%s/answers?order_by=vote_num&page=%d", id, i)
		log.Printf("doSnapAnswer...sid:%d, url:%s", sid, url)
		resp := doPageRequest(url)

		doc, err := goquery.NewDocumentFromResponse(resp)
		if err != nil {
			return err
		}

		doc.Find(".zm-profile-section-list .zm-item").EachWithBreak(func(i int, s *goquery.Selection) bool {
			agree, _ := s.Find(".zm-item-vote .zm-item-vote-count").Attr("data-votecount")

			agreeNum, _ := strconv.Atoi(agree)
			if agreeNum < 250 {
				log.Printf("%d agree, not enough.", agreeNum)
				shouldBreak = true
				return false
			}

			link, _ := s.Find(".question_link").Attr("href")
			link = fmt.Sprintf("http://zhihu.com%s", link)
			l := strings.Split(link, "/")
			answerid := l[len(l)-1]
			title := s.Find(".question_link").Text()

			noshare := 0
			if s.Find(".zm-item-meta .copyright").Text() == " 禁止转载 " {
				noshare = 1
			}

			summary := s.Find(".zh-summary").Text()

			/*d, _ := s.Find(".zm-item-answer").Attr("data-created")
			timestamp, err := strconv.ParseInt(d, 10, 64)
			if err != nil {
				log.Println("strconv.ParseInt failed")
				return false
			}
			date := time.Unix(timestamp, 8)
			*/
			date := time.Now()
			log.Println("date:", date)

			//re, _ := regexp.Compile("(\\n|\\<[\\s\\S]+?\\>)")
			//summary = re.ReplaceAllString(summary, "")
			if len(summary) >= 12 {
				summary = summary[:len(summary)-12]
				if len(summary) > 812 {
					summary = summary[:812]
				}
			}
			length := len(summary)
			ispost := 1
			//log.Printf("sid:%d, id:%s, date:%s, answerid:%s, agree:%s, title:%s, link:%s, ispost, noshare:%d, length:%d, summary:%s",
			//			sid, id, date, answerid, agree, title, link, ispost, noshare, length, string(summary))

			if result, err := stmtTopAnaswer.Exec(sid, id, title, agree, date, answerid, link, ispost, noshare, length, string(summary)); err == nil {
				if lastInsertId, err := result.LastInsertId(); err == nil {
					log.Println("stmtTopAnaswer lastInsertId:", lastInsertId)
				}
			} else {
				log.Println("stmtTopAnaswer err:", err)
				return false
			}
			return true
		})

		if shouldBreak {
			break
		}
	}

	return err
}
