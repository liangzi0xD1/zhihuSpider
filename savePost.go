package main

import (
	"os/exec"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

type QuestionContent struct {
	Title    string
	QuestionId string
	QuestionLink string
	Answers []AnswerContent
}

type AnswerContent struct {
	Answerid string
	Link     string
	Id       string
	Name     string
	Agree    int
	Ispost   int
	Noshare  int
	Length   int
	Summary  string
	Avatar   string
}

func doSavePage() error {
	log.Println("doSavePage...")

	var startAt string
	err := db.QueryRow(`select startAt from snapshot_config where sid = (select max(sid) from snapshot_config where  finished=1)`).Scan(&startAt)
	if err != nil {
		return err
	}

	loc, _ := time.LoadLocation("Local")
	startAtTime, _ := time.Parse("2006-01-02 15:04:05", startAt)
	startAtTime = startAtTime.In(loc)

	hour := startAtTime.Hour()
	log.Println(startAt, startAtTime, loc.String(), hour)

	num := 1
	limit := 32
	yesterdayPage := false
	if hour < 12 {
		num = 2
		yesterdayPage = true
	}

	rows, err := db.Query(`select b.id, u.name, b.title, b.agree, b.answerid, b.link, b.ispost, b.noshare, b.len, b.summary, u.avatar
									from zhihu.usertopanswers a
									INNER JOIN zhihu.usertopanswers b on b.id=a.id and b.answerid = a.answerid and b.sid=a.sid+?
									INNER JOIN zhihu.users u on u.id = a.id
									where a.sid=(select max(sid) from zhihu.snapshot_config where finished=1)-? order by b.agree-a.agree desc limit ?`, num, num, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	daytime := strings.Split(startAtTime.Format(time.RFC3339), "+")[0]
	day := strings.Split(daytime, "T")[0]
	date := strings.Replace(daytime, "T", " ", -1)

	path := "source/_posts/"
	postTitle := "近日精选 "
	fileName := path+day + "-today" + ".md"
	if yesterdayPage {
		postTitle = "昨日精选 "
		fileName = path+day + "-yesterday" + ".md"
	}

	contents := []QuestionContent{}
	for rows.Next() {
		var a AnswerContent
		var q QuestionContent

		rows.Scan(&a.Id, &a.Name, &q.Title, &a.Agree, &a.Answerid, &a.Link, &a.Ispost, &a.Noshare, &a.Length, &a.Summary, &a.Avatar)

		a.Summary = strings.TrimRight(strings.TrimLeft(a.Summary, "\n"), "\n")
		re := regexp.MustCompile("https://.*.zhimg.com")
		a.Avatar = re.ReplaceAllString(a.Avatar, "http://7xojdu.com1.z0.glb.clouddn.com")
		q.QuestionId = strings.Split(a.Link, "/")[len(strings.Split(a.Link, "/"))-3]
		q.QuestionLink = "http://www.zhihu.com/question/"+q.QuestionId

		//log.Printf("id:%s, name:%s, title:%s, agree:%d, answerid:%s, link:%s, ispost:%d, noshare:%d, length:%d, summary:%s, avatar:%s",
		//	a.Id, a.Name, q.Title, a.Agree, a.Answerid, a.Link, a.Ispost, a.Noshare, a.Length, a.Summary, a.Avatar)
		
		isFind := false
		for k, v := range contents {
			if v.QuestionId == q.QuestionId {
				log.Println("found ", v.QuestionId)
				isFind = true
				contents[k].Answers = append(v.Answers, a)
			}
		}

		if !isFind {
			log.Println("not found ", q.QuestionId)
			q.Answers = append(q.Answers, a)
			contents = append(contents, q)
		}
	}

	log.Println(len(contents))
	postPage := fmt.Sprintf("title: %s\n", postTitle + day)
	postPage += fmt.Sprintf("date: %s\n", date)
	postPage += fmt.Sprintf("---\n")
	for k1, v := range contents {
		postPage += fmt.Sprintf("<h3 id='%s'><a href='%s'>%s</a></h3>", v.Title, v.QuestionLink, v.Title)
		postPage += fmt.Sprintf("<div>")
		for k2, a := range v.Answers {
			log.Println("answers:", k1, k2)
			postPage += fmt.Sprintf("<a href='http://zhihu.com/people/%s'>", a.Id)
			postPage += fmt.Sprintf("<img src='%s' align='left'></a>", a.Avatar)
			postPage += fmt.Sprintf("<span>**[%s](http://zhihu.com/people/%s)**:", a.Name, a.Id)
			postPage += fmt.Sprintf(" (*%d*<img class='agreelogo' src='http://7xojdu.com1.z0.glb.clouddn.com/agree.png'>)", a.Agree)
			postPage += fmt.Sprintf(" %s[[阅读全文]](%s)</span>", a.Summary, a.Link)
			postPage += fmt.Sprintf("<div style='clear: both; margin-bottom: 16px;'></div>")
		}
		postPage += fmt.Sprintf("</div>")
		postPage += fmt.Sprintf("\n")
	}

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	file.WriteString(postPage)
	file.Close()

	log.Println("savePage done")
	
	err = deploy()
	if err == nil {
		log.Println("deploy done")
	}
	return err
}

func deploy() error {
	c := exec.Command("hexo", "g")
	err := c.Start()
	if err != nil {
		return err
	}

	log.Printf("Waiting for command to finish...")
	err = c.Wait()
	if err != nil {
		return err
	}
	c = exec.Command("hexo", "g")
	err = c.Start()
	if err != nil {
		return err
	}

	log.Printf("Waiting for command to finish...")
	err = c.Wait()
	if err != nil {
		return err
	}

	c = exec.Command("hexo", "d")
	err = c.Start()
	if err != nil {
		return err
	}

	err = c.Wait()
	if err != nil {
		return err
	}
	
	return err
}