package mail

import (
	"context"
	"errors"
	"log"

	"github.com/Netflix/go-env"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/yoannduc/scoreav-crawler/internal/aws/dynamo"
	"github.com/yoannduc/scoreav-crawler/internal/helpers"
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

func getLastMailNotif(table string, e types.Event) (*types.LastElementNotified, error) {
	// Create the ddb connexion
	svc, err := dynamo.GetConnexion()
	if err != nil {
		return nil, errors.New("Error while creating dynamo connexion : " + err.Error())
	}

	q, err := svc.Query(&dynamodb.QueryInput{
		TableName: aws.String(table),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {S: aws.String("notification")},
			":sk": {S: aws.String("mail#" + e.Source + "#" + e.Type)},
		},
		KeyConditionExpression: aws.String("pk = :pk and begins_with(sk, :sk)"),
		Limit:                  aws.Int64(1),
		ScanIndexForward:       aws.Bool(false),
	})
	if err != nil {
		return nil, errors.New("Error while querying dynamo for last mail notification : " + err.Error())
	}

	var el *types.LastElementNotified
	if *q.Count < 1 {
		return el, nil
	}

	err = dynamodbattribute.Unmarshal(&dynamodb.AttributeValue{M: q.Items[0]}, &el)
	if err != nil {
		return nil, errors.New("Error while decoding last mail notification : " + err.Error())
	}

	return el, nil
}

func getAllSinceLastNotif(table string, e types.Event, le *types.LastElementNotified) ([]*types.Element, error) {
	// Create the ddb connexion
	svc, err := dynamo.GetConnexion()
	if err != nil {
		return nil, errors.New("Error while creating dynamo connexion : " + err.Error())
	}

	qi := &dynamodb.QueryInput{
		TableName: aws.String(table),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {S: aws.String(e.Source + "#" + e.Type)},
		},
		KeyConditionExpression: aws.String("pk = :pk"),
	}

	if le != nil {
		sk, err := le.BuildExclusiveStartKey()
		if err != nil {
			return nil, errors.New("Error building last notification start key : " + err.Error())
		}
		qi.ExclusiveStartKey = sk
	}

	log.Printf("qi | %T | %v\n", qi, qi)

	q, err := svc.Query(qi)
	if err != nil {
		return nil, errors.New("Error while querying dynamo for " + e.Type + " from " + e.Source + " to notify : " + err.Error())
	}
	if *q.Count < 1 {
		return nil, errors.New("No " + e.Type + " from " + e.Source + " since last mail notification")
	}

	var els []*types.Element
	var el *types.Element
	for i := 0; i < len(q.Items); i++ {
		err = dynamodbattribute.Unmarshal(&dynamodb.AttributeValue{M: q.Items[i]}, &el)
		if err != nil {
			return nil, errors.New("Error while decoding last mail notification : " + err.Error())
		}
		els = append(els, el)
		el = nil
	}

	return els, nil
}

func writeLastElem(table string, e types.Event, le *types.Element) error {
	// Create the ddb connexion
	svc, err := dynamo.GetConnexion()
	if err != nil {
		return errors.New("Error while creating dynamo connexion : " + err.Error())
	}

	el, err := dynamodbattribute.MarshalMap(le.ToLastNotify())
	if err != nil {
		return errors.New("Error while encoding last element in notification : " + err.Error())
	}

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      el,
	})
	if err != nil {
		return errors.New("Error writing last mail notification : " + err.Error())
	}

	return nil
}

// HandleRequest handles the lambda job
// It scraps scoreav site to get articles of today which are not yet in ddb
// ctx is the lambda context
// e is the event input
func HandleRequest(ctx context.Context, e types.Event) (string, error) {
	// Set config
	var c config
	if _, err := env.UnmarshalFromEnviron(&c); err != nil {
		return helpers.HandleErr("Error while unwrapping env config : " + err.Error())
	}

	toto, err := getLastMailNotif(c.DdbTable, e)
	if err != nil {
		return helpers.HandleErr(err.Error())
	}

	log.Printf("toto | %T | %v\n", toto, toto)

	titi, err := getAllSinceLastNotif(c.DdbTable, e, toto)
	if err != nil {
		return helpers.HandleErr(err.Error())
	}

	log.Printf("titi | %T | %v\n", titi, titi)

	// log.Printf("titi[0] | %T | %v\n", titi[0], titi[0])

	// log.Printf("titi[3] | %T | %v\n", titi[3], titi[3])

	log.Printf("titi[len(titi) - 1] | %T | %v\n", titi[len(titi)-1], titi[len(titi)-1])

	// TODO Notify SNS

	le := titi[len(titi)-1]
	err = writeLastElem(c.DdbTable, e, le)
	if err != nil {
		return helpers.HandleErr(err.Error())
	}

	return "", nil
}
