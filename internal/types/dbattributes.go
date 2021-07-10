package types

import (
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type LastElementNotified struct {
	Pk     string `json:"pk"`
	Sk     string `json:"sk"`
	LastPk string `json:"last_element_notified_pk"`
	LastSk string `json:"last_element_notified_sk"`
	Date   string `json:"date"`
	Source string `json:"source"`
	Type   string `json:"type"`
}

func (le *LastElementNotified) BuildExclusiveStartKey() (map[string]*dynamodb.AttributeValue, error) {
	return dynamodbattribute.MarshalMap(struct {
		Pk string `json:"pk"`
		Sk string `json:"sk"`
	}{
		Pk: le.LastPk,
		Sk: le.LastSk,
	})
}

type Element struct {
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

func (e *Element) ToLastNotify() *LastElementNotified {
	n := time.Now().Format(time.RFC3339)
	return &LastElementNotified{
		Pk:     "notification",
		Sk:     "mail#" + e.Source + "#" + e.Type + "#" + n,
		LastPk: e.Pk,
		LastSk: e.Sk,
		Date:   n,
		Source: e.Source,
		Type:   e.Type,
	}
}
