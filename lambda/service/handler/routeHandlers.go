package handler

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/pennsieve/model-service-serverless/api/models"
	"github.com/pennsieve/model-service-serverless/api/service"
	"github.com/pennsieve/pennsieve-go-api/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/gateway"
	"log"
)

func getDatasetModelsRoute(s service.GraphService, request events.APIGatewayV2HTTPRequest,
	claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {
	// Create database session and defer closing the session

	// Get the models from Neo4J
	results, err := s.GetDatasetModels(int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId))
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

func postGraphQueryRoute(s service.GraphService, request events.APIGatewayV2HTTPRequest,
	claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {
	apiResponse := events.APIGatewayV2HTTPResponse{}

	parsedRequestBody := models.QueryRequestBody{}
	if err := json.Unmarshal([]byte(request.Body), &parsedRequestBody); err != nil {
		message := "Error: Unable to parse body: " + fmt.Sprint(err)
		apiResponse = events.APIGatewayV2HTTPResponse{
			Body: gateway.CreateErrorMessage(message, 400), StatusCode: 400}
		return &apiResponse, nil
	}

	records, err := s.QueryGraph(parsedRequestBody, int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId))

	if err != nil {
		switch err.(type) {
		case *models.UnknownModelPropertyError:
			apiResponse = events.APIGatewayV2HTTPResponse{
				Body: gateway.CreateErrorMessage(err.Error(), 400), StatusCode: 400}
		case *models.UnknownModelError:
			apiResponse = events.APIGatewayV2HTTPResponse{
				Body: gateway.CreateErrorMessage(err.Error(), 400), StatusCode: 400}
		case *models.UnsupportedOperatorError:
			apiResponse = events.APIGatewayV2HTTPResponse{
				Body: gateway.CreateErrorMessage(err.Error(), 400), StatusCode: 400}
		default:
			log.Println(err)
			apiResponse = events.APIGatewayV2HTTPResponse{
				Body: gateway.CreateErrorMessage("Internal Server Error", 500), StatusCode: 500}
		}

		return &apiResponse, nil

	}

	// CREATING API RESPONSE
	jsonBody, _ := json.Marshal(records)
	apiResponse = events.APIGatewayV2HTTPResponse{Body: string(jsonBody), StatusCode: 200}

	return &apiResponse, nil
}

// postGraphRecordRelationshipRoute creates 1 or more relationships between existing records
func postGraphRecordRelationshipRoute(s service.GraphService, request events.APIGatewayV2HTTPRequest,
	claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {
	apiResponse := events.APIGatewayV2HTTPResponse{}

	parsedRequestBody := models.PostRecordRelationshipRequestBody{}
	if err := json.Unmarshal([]byte(request.Body), &parsedRequestBody); err != nil {
		message := "Error: Unable to parse body: " + fmt.Sprint(err)
		apiResponse = events.APIGatewayV2HTTPResponse{
			Body: gateway.CreateErrorMessage(message, 400), StatusCode: 400}
		return &apiResponse, nil
	}

	response, err := s.CreateRelationships(parsedRequestBody, int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId), claims.UserClaim.NodeId)
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

func postAutocompleteRoute(s service.GraphService, request events.APIGatewayV2HTTPRequest,
	claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {
	apiResponse := events.APIGatewayV2HTTPResponse{}

	parsedRequestBody := models.AutocompleteRequestBody{}
	if err := json.Unmarshal([]byte(request.Body), &parsedRequestBody); err != nil {
		message := "Error: Unable to parse body: " + fmt.Sprint(err)
		apiResponse = events.APIGatewayV2HTTPResponse{
			Body: gateway.CreateErrorMessage(message, 400), StatusCode: 400}
		return &apiResponse, nil
	}

	values, err := s.Autocomplete(parsedRequestBody, int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId))

	if err != nil {
		switch err.(type) {
		case *models.UnknownModelPropertyError:
			apiResponse = events.APIGatewayV2HTTPResponse{
				Body: gateway.CreateErrorMessage(err.Error(), 400), StatusCode: 400}
		case *models.UnknownModelError:
			apiResponse = events.APIGatewayV2HTTPResponse{
				Body: gateway.CreateErrorMessage(err.Error(), 400), StatusCode: 400}
		case *models.UnsupportedOperatorError:
			apiResponse = events.APIGatewayV2HTTPResponse{
				Body: gateway.CreateErrorMessage(err.Error(), 400), StatusCode: 400}
		default:
			log.Println(err)
			apiResponse = events.APIGatewayV2HTTPResponse{
				Body: gateway.CreateErrorMessage("Internal Server Error", 500), StatusCode: 500}
		}

		return &apiResponse, nil

	}

	// CREATING API RESPONSE
	responseBody := models.AutocompleteResponse{
		ModelName: parsedRequestBody.Model,
		Property:  parsedRequestBody.Property,
		Values:    values,
	}

	jsonBody, _ := json.Marshal(responseBody)
	apiResponse = events.APIGatewayV2HTTPResponse{Body: string(jsonBody), StatusCode: 200}

	return &apiResponse, nil
}
