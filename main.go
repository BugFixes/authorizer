package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bugfixes/authorizer/service"
)

func main() {
	lambda.Start(service.Handler)
}
