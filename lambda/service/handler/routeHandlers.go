package handler

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/pennsieve/model-service-serverless/api/models"
	"github.com/pennsieve/model-service-serverless/api/store"
	"github.com/pennsieve/pennsieve-go-api/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/gateway"
)

func getDatasetModelsRoute(s store.GraphStore, request events.APIGatewayV2HTTPRequest, claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {
	// Create database session and defer closing the session

	// Get the models from Neo4J
	results, err := s.GetModels(int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId))
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

func postGraphQueryRoute(s store.GraphStore, request events.APIGatewayV2HTTPRequest, claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {
	apiResponse := events.APIGatewayV2HTTPResponse{}

	parsedRequestBody := models.QueryRequestBody{}
	if err := json.Unmarshal([]byte(request.Body), &parsedRequestBody); err != nil {
		message := "Error: Unable to parse body: " + fmt.Sprint(err)
		apiResponse = events.APIGatewayV2HTTPResponse{
			Body: gateway.CreateErrorMessage(message, 400), StatusCode: 400}
		return &apiResponse, nil
	}

	err := s.Query(int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId), parsedRequestBody)

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
func postGraphRecordRelationshipRoute(s store.GraphStore, request events.APIGatewayV2HTTPRequest, claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {
	apiResponse := events.APIGatewayV2HTTPResponse{}

	parsedRequestBody := models.PostRecordRelationshipRequestBody{}
	if err := json.Unmarshal([]byte(request.Body), &parsedRequestBody); err != nil {
		message := "Error: Unable to parse body: " + fmt.Sprint(err)
		apiResponse = events.APIGatewayV2HTTPResponse{
			Body: gateway.CreateErrorMessage(message, 400), StatusCode: 400}
		return &apiResponse, nil
	}

	response, err := s.CreateRelationShips(int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId), claims.UserClaim.NodeId, parsedRequestBody)
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
