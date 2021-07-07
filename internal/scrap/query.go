package scrap

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/yoannduc/scoreav-crawler/internal/types"
)

func queryFullPage(l string) (string, error) {
	// Request the HTML page.
	res, err := http.Get(l)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if 200 < res.StatusCode || res.StatusCode >= 300 {
		return "", errors.New("Querying " + l + " returned status of " + strconv.Itoa(res.StatusCode))
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	var rs []string
	doc.Find(".cb-entry-content>p[style*=\"text-align: justify;\"]").Each(func(_ int, s *goquery.Selection) {
		rs = append(rs, s.Text())
	})

	return strings.Join(rs, "\n\n"), nil
}

func retrieveElem(s *goquery.Selection, ev types.Event, c chan<- *types.Element) {
	id := uuid.NewString()
	d, _ := s.Find(".cb-date>time").Attr("datetime")

	// TODO Check how to close this routine early if no link retrieved
	l, ok := s.Find(".cb-post-title>a").Attr("href")
	var ld string
	var err error
	if ok {
		ld, err = queryFullPage(l)
		if err != nil {
			ld = ""
			log.Println(err)
		}
	}

	c <- &types.Element{
		Pk:     ev.Source,
		Sk:     ev.Type + "#" + l,
		ID:     id,
		Link:   l,
		Title:  s.Find(".cb-post-title").Text(),
		SDesc:  s.Find(".cb-excerpt").Text(),
		LDesc:  ld,
		Date:   d,
		Source: ev.Source,
		Type:   ev.Type,
	}
}

func buildElemList(d *goquery.Document, ev types.Event) []*types.Element {
	var el []*types.Element

	switch ev.Source {
	case "scoreav":
		// Counter to know how many channel output to wait for
		n := 0
		c := make(chan *types.Element, 1)
		defer close(c)

		d.Find("article").Each(func(_ int, s *goquery.Selection) {
			go retrieveElem(s, ev, c)

			n++
		})

		for i := 0; i < n; i++ {
			e := <-c
			el = append(el, e)
		}
	}

	return el
}
