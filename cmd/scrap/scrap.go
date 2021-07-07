package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/yoannduc/scoreav-crawler/internal/scrap"
)

func main() {
	lambda.Start(scrap.HandleRequest)
}
