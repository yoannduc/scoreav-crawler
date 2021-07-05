package dynamo

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/yoannduc/scoreav-crawler/internal/aws/session"
)

var svc *dynamodb.DynamoDB

// GetConnexion returns an instance of dynamodb connexion defined by session
func GetConnexion() (*dynamodb.DynamoDB, error) {
	if svc != nil {
		return svc, nil
	}

	// AWS session
	sess, err := session.GetSession()
	if err != nil {
		return nil, err
	}

	svc = dynamodb.New(sess)

	return svc, nil
}
