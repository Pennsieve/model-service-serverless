package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/pennsieve/model-service-serverless/api/models"
	"github.com/pennsieve/model-service-serverless/api/query"
	"github.com/pennsieve/model-service-serverless/api/records"
	"github.com/pennsieve/pennsieve-go-api/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/gateway"
)

func getDatasetModelsRoute(request events.APIGatewayV2HTTPRequest, claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {
	// Create database session and defer closing the session
	session := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(context.Background())

	// Get the models from Neo4J
	results, err := models.GetModels(session, int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId))
	if err != nil {
		return nil, err
	}

	// Parse response into JSON structure
	jsonBody, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}
	apiResponse := events.APIGatewayV2HTTPResponse{Body: string(jsonBody), StatusCode: 200}
	return &apiResponse, nil
}

func postGraphQueryRoute(request events.APIGatewayV2HTTPRequest, claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {
	apiResponse := events.APIGatewayV2HTTPResponse{}

	parsedRequestBody := query.QueryRequestBody{}
	if err := json.Unmarshal([]byte(request.Body), &parsedRequestBody); err != nil {
		message := "Error: Unable to parse body: " + fmt.Sprint(err)
		apiResponse = events.APIGatewayV2HTTPResponse{
			Body: gateway.CreateErrorMessage(message, 400), StatusCode: 400}
		return &apiResponse, nil
	}

	session := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(context.Background())

	err := query.Query(session, int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId), parsedRequestBody)

	if err != nil {
		return nil, err
	}

	// CREATING API RESPONSE
	responseBody := "Success"
	jsonBody, _ := json.Marshal(responseBody)
	apiResponse = events.APIGatewayV2HTTPResponse{Body: string(jsonBody), StatusCode: 200}

	return &apiResponse, nil
}

// postGraphRecordRelationshipRoute creates 1 or more relationships between existing records
func postGraphRecordRelationshipRoute(request events.APIGatewayV2HTTPRequest, claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {
	apiResponse := events.APIGatewayV2HTTPResponse{}

	parsedRequestBody := records.PostRecordRelationshipRequestBody{}
	if err := json.Unmarshal([]byte(request.Body), &parsedRequestBody); err != nil {
		message := "Error: Unable to parse body: " + fmt.Sprint(err)
		apiResponse = events.APIGatewayV2HTTPResponse{
			Body: gateway.CreateErrorMessage(message, 400), StatusCode: 400}
		return &apiResponse, nil
	}

	session := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(context.Background())

	response, err := records.CreateRelationShips(session, int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId), claims.UserClaim.NodeId, parsedRequestBody)
	if err != nil {
		message := err.Error()
		apiResponse = events.APIGatewayV2HTTPResponse{
			Body: gateway.CreateErrorMessage(message, 400), StatusCode: 400}
		return &apiResponse, nil

		return nil, err
	}

	// CREATING API RESPONSE
	jsonBody, _ := json.Marshal(response)
	apiResponse = events.APIGatewayV2HTTPResponse{Body: string(jsonBody), StatusCode: 200}

	return &apiResponse, nil
}
