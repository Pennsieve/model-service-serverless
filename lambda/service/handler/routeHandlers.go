package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/pennsieve/model-service-serverless/api/models"
	"log"
)

func getDatasetModelsRoute(request events.APIGatewayV2HTTPRequest, claims *Claims) (*events.APIGatewayV2HTTPResponse, error) {
	fmt.Println("Handling GET /graph/models")

	datasetId := 1715
	organizationId := 19

	session := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(context.Background())

	models.getModels(session, datasetId, organizationId)

	transaction, err := session.BeginTransaction(context.Background())
	if err != nil {
		log.Println(err)
		return nil, err
	}

	result, err := transaction.Run(context.Background(),
		"CREATE (a:Greeting) SET a.message = $message RETURN a.message + ', from node ' + id(a)",
		map[string]any{"message": "hello, world"})

	if err != nil {
		return nil, err
	}

	fmt.Println(result)

	apiResponse := events.APIGatewayV2HTTPResponse{}

	// CREATING API RESPONSE
	responseBody := "Hello there."

	jsonBody, _ := json.Marshal(responseBody)
	apiResponse = events.APIGatewayV2HTTPResponse{Body: string(jsonBody), StatusCode: 200}

	return &apiResponse, nil
}

func postGraphQueryRoute(request events.APIGatewayV2HTTPRequest, claims *Claims) (*events.APIGatewayV2HTTPResponse, error) {
	fmt.Println("Handling POST /graph/query request")

	session := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(context.Background())

	transaction, err := session.BeginTransaction(context.Background())
	if err != nil {
		log.Println(err)
		return nil, err
	}

	result, err := transaction.Run(context.Background(),
		"CREATE (a:Greeting) SET a.message = $message RETURN a.message + ', from node ' + id(a)",
		map[string]any{"message": "hello, world"})

	if err != nil {
		return nil, err
	}

	fmt.Println(result)

	apiResponse := events.APIGatewayV2HTTPResponse{}

	// CREATING API RESPONSE
	responseBody := "Hello there."

	jsonBody, _ := json.Marshal(responseBody)
	apiResponse = events.APIGatewayV2HTTPResponse{Body: string(jsonBody), StatusCode: 200}

	return &apiResponse, nil
}
