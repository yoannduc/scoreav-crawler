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
	Pk     string `json:"pk"`
	Sk     string `json:"sk"`
	ID     string `json:"id"`
	Link   string `json:"link"`
	Title  string `json:"title"`
	SDesc  string `json:"short_description"`
	LDesc  string `json:"long_description,omitempty"`
	Date   string `json:"date"`
	Source string `json:"source"`
	Type   string `json:"type"`
}

type event struct {
	Source string `json:"source"`
	Type   string `json:"type"`
}

func buildURI(e event) (string, error) {
	switch e.Source {
	case "scoreav":
		switch e.Type {
		case "news":
			return "http://www.scoreav.com/news/", nil
		case "focus":
			return "http://www.scoreav.com/chronique/", nil
		default:
			return "", errors.New("Type empty or not handled for source " + e.Source + " (inputed: " + e.Type + ")")
		}
	default:
		return "", errors.New("Source not handled or empty (inputed:\"" + e.Source + "\")")
	}
}

func handleErr(e string) (string, error) {
	log.Println(e)
	return "", errors.New(e)
}

func buildElemList(d *goquery.Document, ev event) []*element {
	var el []*element

	switch ev.Source {
	case "scoreav":
		d.Find("article").Each(func(_ int, s *goquery.Selection) {
			id := uuid.NewString()
			l, _ := s.Find(".cb-post-title>a").Attr("href")
			d, _ := s.Find(".cb-date>time").Attr("datetime")

			e := &element{
				Pk:     ev.Source,
				Sk:     ev.Type + "#" + id,
				ID:     id,
				Link:   l,
				Title:  s.Find(".cb-post-title").Text(),
				SDesc:  s.Find(".cb-excerpt").Text(),
				Date:   d,
				Source: ev.Source,
				Type:   ev.Type,
			}

			el = append(el, e)
		})
	}

	return el
}

// HandleRequest handles the lambda job
// It scraps scoreav site to get articles of today which are not yet in ddb
// ctx is the lambda context
// e is the event input
func HandleRequest(ctx context.Context, e event) (string, error) {
	uri, err := buildURI(e)
	if err != nil {
		return handleErr(err.Error())
	}

	resp, err := http.Get(uri)
	if err != nil {
		return handleErr(err.Error())
	}
	defer resp.Body.Close()

	d, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return handleErr(err.Error())
	}

	el := buildElemList(d, e)

	json, err := jsoniter.MarshalToString(el)
	if err != nil {
		return handleErr(err.Error())
	}

	log.Println(json)

	return json, nil
}

func main() {
	lambda.Start(HandleRequest)
}
