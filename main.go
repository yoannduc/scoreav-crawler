package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	jsoniter "github.com/json-iterator/go"
)

type element struct {
	Link  string `json:"link"`
	Title string `json:"title"`
	Date  string `json:"date"`
}

func main() {
	fmt.Println("Hello World !")

	var el []*element

	resp, err := http.Get("http://www.scoreav.com/news/")
	if err != nil {
		fmt.Println(errors.New("lul 1"))
	}
	defer resp.Body.Close()

	d, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		fmt.Println(errors.New("lul 2"))
	}

	fmt.Println(d)

	d.Find("article").Each(func(_ int, s *goquery.Selection) {
		l, _ := s.Find(".cb-post-title>a").Attr("href")
		d, _ := s.Find(".cb-date>time").Attr("datetime")

		e := &element{Link: l, Title: s.Find(".cb-post-title").Text(), Date: d}

		fmt.Println(e)

		el = append(el, e)
	})

	fmt.Println(el)

	json, err := jsoniter.MarshalToString(el)
	if err != nil {
		fmt.Println(errors.New("lul 3"))
	}

	fmt.Println(json)
}
