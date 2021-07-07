package scrap

import (
	"errors"
	"log"
	"math"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/yoannduc/scoreav-crawler/internal/aws/dynamo"
	"github.com/yoannduc/scoreav-crawler/internal/types"
)

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
