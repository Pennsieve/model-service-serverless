package handler

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func postGraphQueryRoute(request events.APIGatewayV2HTTPRequest, claims *Claims) (*events.APIGatewayV2HTTPResponse, error) {
	fmt.Println("Handling POST /graph/query request")

	dbUri := "bolt-://10.11.1.51:7687"
	driver, err := neo4j.NewDriver(dbUri, neo4j.BasicAuth("model_service_user", "8RY@BheMwUvNZ3", ""))
	if err != nil {
		panic(err)
	}
	// Handle driver lifetime based on your application lifetime requirements  driver's lifetime is usually
	// bound by the application lifetime, which usually implies one driver instance per application
	defer driver.Close()

	session := driver.NewSession(neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	greeting, err := session.WriteTransaction(func(transaction neo4j.Transaction) (any, error) {
		result, err := transaction.Run(
			"CREATE (a:Greeting) SET a.message = $message RETURN a.message + ', from node ' + id(a)",
			map[string]any{"message": "hello, world"})
		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})
	if err != nil {
		return nil, err
	}
	fmt.Println(greeting)

	defer session.Close()

	apiResponse := events.APIGatewayV2HTTPResponse{}

	// CREATING API RESPONSE
	responseBody := "Hello there."

	jsonBody, _ := json.Marshal(responseBody)
	apiResponse = events.APIGatewayV2HTTPResponse{Body: string(jsonBody), StatusCode: 200}

	return &apiResponse, nil
}
