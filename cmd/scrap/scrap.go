package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
)

type element struct {
	ID    string `json:"id"`
	Link  string `json:"link"`
	Title string `json:"title"`
	Date  string `json:"date"`
}

type event struct{}

// HandleRequest handles the lambda job
// It scraps scoreav site to get articles of today which are not yet in ddb
// ctx is the lambda context
// e is the event input
func HandleRequest(ctx context.Context, e event) (string, error) {
	log.Println("Hello World !")

	var el []*element

	resp, err := http.Get("http://www.scoreav.com/news/")
	// resp, err := http.Get("http://www.scoreav.com/chronique/")
	if err != nil {
		log.Println(errors.New("lul 1"))
		return "", err
	}
	defer resp.Body.Close()

	d, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Println(errors.New("lul 2"))
		return "", err
	}

	log.Println(d)

	d.Find("article").Each(func(_ int, s *goquery.Selection) {
		l, _ := s.Find(".cb-post-title>a").Attr("href")
		d, _ := s.Find(".cb-date>time").Attr("datetime")

		e := &element{ID: uuid.NewString(), Link: l, Title: s.Find(".cb-post-title").Text(), Date: d}

		log.Println(e)

		el = append(el, e)
	})

	log.Println(el)

	json, err := jsoniter.MarshalToString(el)
	if err != nil {
		log.Println(errors.New("lul 3"))
		return "", err
	}

	log.Println(json)

	return "", nil
}

func main() {
	lambda.Start(HandleRequest)
}
