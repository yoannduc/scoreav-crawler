package helpers

import (
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/yoannduc/scoreav-crawler/internal/aws/dynamo"
	"github.com/yoannduc/scoreav-crawler/internal/types"
)

// HandleErr handles error for lambda functions.
// It logs the error and returns the correct lambda handler format
// e is the error string to log and return
func HandleErr(e string) (string, error) {
	log.Println(e)
	return "", errors.New(e)
}

// GetLastUpdatedDate gets the last updated date
// table is the table name
// e is the lambda event
func GetLastUpdatedDate(table string, e types.Event) (string, error) {
	// Create the ddb connexion
	svc, err := dynamo.GetConnexion()
	if err != nil {
		return "", errors.New("Error while creating dynamo connexion : " + err.Error())
	}

	q, err := svc.Query(&dynamodb.QueryInput{
		TableName: aws.String(table),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {S: aws.String(e.Source + "#" + e.Type)},
		},
		ExpressionAttributeNames: map[string]*string{
			"#date": aws.String("date"),
		},
		KeyConditionExpression: aws.String("pk = :pk"),
		ProjectionExpression:   aws.String("#date"),
		Limit:                  aws.Int64(1),
		ScanIndexForward:       aws.Bool(false),
	})
	if err != nil {
		return "", errors.New("Error while querying dynamo for last update : " + err.Error())
	}

	if *q.Count < 1 || q.Items[0]["date"] == nil || q.Items[0]["date"].NULL != nil {
		return "", errors.New("Empty last date")
	}

	return *q.Items[0]["date"].S, nil
}
