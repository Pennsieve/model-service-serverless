package service

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pennsieve/model-service-serverless/service/handler"
)

func main() {
	lambda.Start(handler.ModelServiceHandler)
}
