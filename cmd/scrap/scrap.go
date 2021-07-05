package main

import (
	"context"
	"errors"
	"log"
	"math"
	"net/http"
	"time"

	env "github.com/Netflix/go-env"
	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/yoannduc/scoreav-crawler/internal/aws/dynamo"
)

type config struct {
	DdbTable string `env:"AWS_DDB_TABLE_NAME,required=true"`
}

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

// TODO Add some logs
// TODO Think about db keys with uniqueness in mind

// HandleRequest handles the lambda job
// It scraps scoreav site to get articles of today which are not yet in ddb
// ctx is the lambda context
// e is the event input
func HandleRequest(ctx context.Context, e event) (string, error) {
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

	// Create the ddb connexion
	svc, err := dynamo.GetConnexion()
	if err != nil {
		return handleErr("Error while creating dynamo connexion : " + err.Error())
	}

	// Array of ddb objects to write
	ddbWrites := make([]*dynamodb.WriteRequest, 0)

	for _, el := range els {
		da, err := dynamodbattribute.MarshalMap(el)
		if err != nil {
			// TODO add elem in err
			return handleErr("Error while marshalling into dynamo format: " + err.Error())
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
	done := make(chan bool)

	for i := 0; i < batches; i++ {
		// Define the start and end of this batch
		// If end is more than full length, get the length as cannot take more than cap
		start, end := i*maxWriteOps, (i+1)*maxWriteOps
		if end > len(ddbWrites) {
			end = len(ddbWrites)
		}

		log.Printf("BATCH WRITE | batches: %v | i: %v | start: %v | end: %v\n", batches, i, start, end)

		go func(ddbn string, w []*dynamodb.WriteRequest, c chan<- bool) {
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
		}(c.DdbTable, ddbWrites[start:end], done)

	}

	for i := 0; i < batches; i++ {
		d := <-done
		if !d {
			return handleErr("Error while writing to dynamodb. See the logs for more informations")
		}
	}

	json, err := jsoniter.MarshalToString(els)
	if err != nil {
		return handleErr(err.Error())
	}

	log.Println(json)

	return json, nil
}

func main() {
	lambda.Start(HandleRequest)
}
