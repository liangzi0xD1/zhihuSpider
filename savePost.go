package main

import (
	"html/template"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

type PageContent struct {
	Id       string
	Name     string
	Sid      int
	Title    string
	Agree    int
	Answerid string
	Link     string
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

	rows, err := db.Query(`select b.id, u.name, b.sid, b.title, b.agree-a.agree agree, b.answerid, b.link, b.ispost, b.noshare, b.len, b.summary, u.avatar
									from zhihu.usertopanswers a
									INNER JOIN zhihu.usertopanswers b on b.id=a.id and b.answerid = a.answerid and b.sid=a.sid+?
									INNER JOIN zhihu.users u on u.id = a.id
									where b.sid=(select max(sid) from zhihu.snapshot_config where finished=1) order by agree desc limit ?`, num, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	day := strings.Split(startAtTime.Format(time.RFC3339), "T")[0]

	postTitle := "近日精选 "
	fileName := day + "-today" + ".md"
	if yesterdayPage {
		postTitle = "昨日精选 "
		fileName = day + "-yesterday" + ".md"
	}

	const tpl = `title: {{.PTitle}}
date: {{.Date}}
tags:
---

<ul>
{{range .Contents}}<li><div style="height:160px"><h3><a href="{{.Link}}">{{.Title}}</a></h3><p><a href="http://zhihu.com/people/{{.Id}}"><img src="{{.Avatar}}" align="left"></a><span style="display:block;padding-left:120px">**[{{.Name}}](http://zhihu.com/people/{{.Id}})**: (*{{.Agree}}* 新增赞同){{.Summary}}</span></p></div></li>{{end}}
</ul>`

	t, err := template.New("post").Parse(tpl)
	if err != nil {
		return err
	}

	contents := []PageContent{}
	for rows.Next() {
		var c PageContent

		rows.Scan(&c.Id, &c.Name, &c.Sid, &c.Title, &c.Agree, &c.Answerid, &c.Link, &c.Ispost, &c.Noshare, &c.Length, &c.Summary, &c.Avatar)

		log.Printf("id:%s, name:%s, sid:%d, title:%s, agree:%d, answerid:%s, link:%s, ispost:%d, noshare:%d, length:%d, summary:%s, avatar:%s",
			c.Id, c.Name, c.Sid, c.Title, c.Agree, c.Answerid, c.Link, c.Ispost, c.Noshare, c.Length, c.Summary, c.Avatar)
		c.Summary = strings.TrimRight(strings.TrimLeft(c.Summary, "\n"), "\n")
		re := regexp.MustCompile("https://.*.zhimg.com")
		c.Avatar = re.ReplaceAllString(c.Avatar, "http://avatar.wanzhihu.com")
		contents = append(contents, c)
	}

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	type Data struct {
		PTitle   string
		Date     string
		Contents []PageContent
	}
	var data Data
	data.PTitle = postTitle + day
	data.Date = day
	data.Contents = contents

	err = t.Execute(file, data)
	if err != nil {
		return err
	}

	log.Println("savePage done")
	return err
}
