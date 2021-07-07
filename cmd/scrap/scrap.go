package main

import (
	"context"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	env "github.com/Netflix/go-env"
	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/yoannduc/scoreav-crawler/internal/aws/dynamo"
	"github.com/yoannduc/scoreav-crawler/internal/types"
)

type config struct {
	DdbTable string `env:"AWS_DDB_TABLE_NAME,required=true"`
	DdbSave  bool   `env:"SAVE_TO_DDB,default=true"`
}

type output struct {
	types.Event
	Processed int `json:"processed"`
}

func buildURI(e types.Event) (string, error) {
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

func writeBatch(svc *dynamodb.DynamoDB, ddbn string, w []*dynamodb.WriteRequest, c chan<- bool) {
	var u map[string][]*dynamodb.WriteRequest
	r := 0
	for {
		if len(u) < 1 {
			u = map[string][]*dynamodb.WriteRequest{
				ddbn: w,
			}
		}
		result, err := svc.BatchWriteItem(&dynamodb.BatchWriteItemInput{
			RequestItems: u,
		})
		if err != nil {
			log.Println("Error while writing to dynamodb: " + err.Error())
			c <- false
			break
		}

		if len(result.UnprocessedItems) < 1 {
			if r > 0 {
				log.Println("Retry went well")
			}
			c <- true
			break
		}

		log.Println("Batch write went without errors, but there are still  some elements to write, retrying in a while")

		if r > 3 {
			log.Println("Too many retries, halting this batch write")
			c <- false
			break
		}

		u = result.UnprocessedItems

		r++
		time.Sleep(time.Duration(int64(math.Floor((math.Pow(2, float64(r))-1)*0.5))) * time.Second)
	}
}

func writeToDdb(els []*types.Element, table string) error {
	// Create the ddb connexion
	svc, err := dynamo.GetConnexion()
	if err != nil {
		return errors.New("Error while creating dynamo connexion : " + err.Error())
	}

	// Array of ddb objects to write
	ddbWrites := make([]*dynamodb.WriteRequest, 0)

	for _, el := range els {
		da, err := dynamodbattribute.MarshalMap(el)
		if err != nil {
			// TODO add elem in err
			return errors.New("Error while marshalling into dynamo format: " + err.Error())
		}

		ddbWrites = append(ddbWrites, &dynamodb.WriteRequest{
			PutRequest: &dynamodb.PutRequest{
				Item: da,
			},
		})
	}

	// Write can only be done by batches of 25 max
	// Calculate how many batches the whole write should need
	maxWriteOps := 25
	batches := int(math.Ceil(float64(len(ddbWrites)) / float64(maxWriteOps)))
	done := make(chan bool, 1)
	defer close(done)

	for i := 0; i < batches; i++ {
		// Define the start and end of this batch
		// If end is more than full length, get the length as cannot take more than cap
		start, end := i*maxWriteOps, (i+1)*maxWriteOps
		if end > len(ddbWrites) {
			end = len(ddbWrites)
		}

		log.Printf("BATCH WRITE | batches: %v | i: %v | start: %v | end: %v\n", batches, i, start, end)

		go writeBatch(svc, table, ddbWrites[start:end], done)
	}

	for i := 0; i < batches; i++ {
		d := <-done
		if !d {
			return errors.New("Error while writing to dynamodb. See the logs for more informations")
		}
	}

	return nil
}

// TODO Add some logs
// TODO Think about db keys with uniqueness in mind

// HandleRequest handles the lambda job
// It scraps scoreav site to get articles of today which are not yet in ddb
// ctx is the lambda context
// e is the event input
func HandleRequest(ctx context.Context, e types.Event) (string, error) {
	// Set config
	var c config
	if _, err := env.UnmarshalFromEnviron(&c); err != nil {
		return handleErr("Error while unwrapping env config : " + err.Error())
	}

	uri, err := buildURI(e)
	if err != nil {
		return handleErr("Error building the uri: " + err.Error())
	}

	resp, err := http.Get(uri)
	if err != nil {
		return handleErr("Error while querying the endpoint :" + err.Error() + " (endpoint: " + uri + ")")
	}
	defer resp.Body.Close()

	d, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return handleErr("Error while decoding response into goquery: " + err.Error())
	}

	els := buildElemList(d, e)

	// If we do not want to save to db (to debug scrapping), return scrapped list
	if !c.DdbSave {
		j, err := jsoniter.MarshalToString(els)
		if err != nil {
			return handleErr(err.Error())
		}

		log.Println(j)

		return j, nil
	}

	err = writeToDdb(els, c.DdbTable)
	if err != nil {
		return handleErr(err.Error())
	}

	j, err := jsoniter.MarshalToString(&output{
		Processed: len(els),
		Event:     e,
	})
	if err != nil {
		return handleErr(err.Error())
	}

	log.Println(j)

	return j, nil
}

func main() {
	lambda.Start(HandleRequest)
}
