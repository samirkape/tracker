package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	tracker "github.com/samirkape/tracker"
)

type MyEvent struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (string, error) {
	tracker.Track()
	return "", nil
}

func main() {
	lambda.Start(HandleRequest)
}
