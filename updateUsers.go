package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func getNewUsers() {

	rows, err := db.Query("select id, name from users")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var id, name string

		rows.Scan(&id, &name)
		log.Println(id, name)
		getUserInfo(id, name)
	}
	rows.Close()
}

func getUserInfo(id, name string) {
	resp := doPageRequest("http://www.zhihu.com/people/" + id + "/followees")
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("INSERT INTO users(id, name) VALUES(?,?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	var follower, agree int
	doc.Find(".zm-profile-card .zm-list-content-medium").Each(func(i int, s *goquery.Selection) {
		name := s.Find("h2").Text()
		href, exist := s.Find(".zg-link").Attr("href")
		if exist != true {
			panic(err)
		}
		h := strings.Split(href, "/")
		id := h[len(h)-1]

		s.Find(".zg-link-gray-normal").Each(func(i int, s *goquery.Selection) {
			p := strings.Split(s.Text(), " ")
			if i == 0 {
				follower, _ = strconv.Atoi(p[0])
			} else if i == 3 {
				agree, _ = strconv.Atoi(p[0])
			}
		})
		log.Printf("follower:%d, agree:%d", follower, agree)

		if follower >= 1000 && agree >= 1000 {

			if result, err := stmt.Exec(id, name); err == nil {
				if id, err := result.LastInsertId(); err == nil {
					log.Println("insert id : ", id)
				}
			} else {
				log.Println("err:", err)
			}
		} else {
			if result, err := db.Exec("DELETE * FROM users WHERE id=? and name=?", id, name); err == nil {
				if id, err := result.LastInsertId(); err == nil {
					log.Println("delete id :", id)
				}
			} else {
				log.Println("err:", err)
			}
		}
	})
}
