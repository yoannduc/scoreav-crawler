package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/yoannduc/scoreav-crawler/internal/mail"
)

func main() {
	lambda.Start(mail.HandleRequest)
}
