package main

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
)

func doPageRequest(url string) *http.Response {
	client := &http.Client{}

	cookie := "z_c0=QUJBTWVwa0dfd2dYQUFBQVlRSlZUZGpBY0ZZMTljMUktSDJkclJ0VFZsaUxUSm1XRTBpOWNRPT0=|1447637976|da89e3b90ac1b614221cd48e2832e20e04c0949b"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.80 Safari/537.36")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Host", "www.zhihu.com")

	resp, err := client.Do(req)
	return resp
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("mysql", "root:y1w2j35217@tcp(localhost:3306)/zhihu?charset=utf8")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(100)
	defer db.Close()

	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("os.Open failed, err:", err)
	}
	defer f.Close()

	w := io.MultiWriter(f, os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(w)

	log.Println("go working...at", time.Now())

	snapUser()
	log.Println("creating cron task...")
	c := cron.New()
	c.AddFunc("0 12 8,20 * * * ", func() {
		snapUser()
	})
	c.Start()

	log.Println(http.ListenAndServe("0.0.0.0:4000", nil))
}
