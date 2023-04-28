package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/pennsieve/model-service-serverless/api/models"
	"github.com/pennsieve/model-service-serverless/api/models/query"
	"github.com/pennsieve/model-service-serverless/api/store"
	"github.com/pennsieve/pennsieve-go-core/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/gateway"
	log "github.com/sirupsen/logrus"
)

func getDatasetModelsRoute(s *store.ModelServiceStore, request events.APIGatewayV2HTTPRequest,
	claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {
	// Create database session and defer closing the session

	// Get the models from Neo4J
	ctx := context.Background()
	results, err := s.GetDatasetModels(ctx, int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId))
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

func postGraphQueryRoute(s *store.ModelServiceStore, request events.APIGatewayV2HTTPRequest,
	claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {
	apiResponse := events.APIGatewayV2HTTPResponse{}

	parsedRequestBody := query.QueryRequestBody{}
	if err := json.Unmarshal([]byte(request.Body), &parsedRequestBody); err != nil {
		message := "Error: Unable to parse body: " + fmt.Sprint(err)
		apiResponse = events.APIGatewayV2HTTPResponse{
			Body: gateway.CreateErrorMessage(message, 400), StatusCode: 400}
		return &apiResponse, nil
	}

	ctx := context.Background()

	response, err := s.QueryGraph(ctx, parsedRequestBody, int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId))

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
	jsonBody, _ := json.Marshal(response)
	apiResponse = events.APIGatewayV2HTTPResponse{Body: string(jsonBody), StatusCode: 200}

	return &apiResponse, nil
}

// postGraphRecordRelationshipRoute creates 1 or more relationships between existing records
func postGraphRecordRelationshipRoute(s *store.ModelServiceStore, request events.APIGatewayV2HTTPRequest,
	claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {
	apiResponse := events.APIGatewayV2HTTPResponse{}

	parsedRequestBody := models.PostRecordRelationshipRequestBody{}
	if err := json.Unmarshal([]byte(request.Body), &parsedRequestBody); err != nil {
		message := "Error: Unable to parse body: " + fmt.Sprint(err)
		apiResponse = events.APIGatewayV2HTTPResponse{
			Body: gateway.CreateErrorMessage(message, 400), StatusCode: 400}
		return &apiResponse, nil
	}

	ctx := context.Background()

	response, err := s.CreateRelationships(ctx, parsedRequestBody, int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId), claims.UserClaim.NodeId)
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

func postAutocompleteRoute(s *store.ModelServiceStore, request events.APIGatewayV2HTTPRequest,
	claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {
	apiResponse := events.APIGatewayV2HTTPResponse{}

	parsedRequestBody := query.AutocompleteRequestBody{}
	if err := json.Unmarshal([]byte(request.Body), &parsedRequestBody); err != nil {
		message := "Error: Unable to parse body: " + fmt.Sprint(err)
		apiResponse = events.APIGatewayV2HTTPResponse{
			Body: gateway.CreateErrorMessage(message, 400), StatusCode: 400}
		return &apiResponse, nil
	}

	ctx := context.Background()

	values, err := s.Autocomplete(ctx, parsedRequestBody, int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId))

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
	responseBody := query.AutocompleteResponse{
		ModelName: parsedRequestBody.Model,
		Property:  parsedRequestBody.Property,
		Values:    values,
	}

	jsonBody, _ := json.Marshal(responseBody)
	apiResponse = events.APIGatewayV2HTTPResponse{Body: string(jsonBody), StatusCode: 200}

	return &apiResponse, nil
}

func getMetaDataForPackage(s *store.ModelServiceStore, request events.APIGatewayV2HTTPRequest,
	claims *authorizer.Claims) (*events.APIGatewayV2HTTPResponse, error) {

	maxDepth := 10
	apiResponse := events.APIGatewayV2HTTPResponse{}
	queryParams := request.QueryStringParameters

	// Get Manifest ID
	var packageId string
	var found bool
	if packageId, found = queryParams["package_id"]; !found {
		message := "Error: Package ID not specified"
		apiResponse = events.APIGatewayV2HTTPResponse{
			Body: gateway.CreateErrorMessage(message, 400), StatusCode: 400}
		return &apiResponse, nil
	}

	ctx := context.Background()
	records, err := s.GetRecordsForPackage(ctx, int(claims.DatasetClaim.IntId), int(claims.OrgClaim.IntId), packageId, maxDepth)
	if err != nil {
		message := fmt.Sprintf("Error: Could not get metadata for package: %v", err)
		apiResponse = events.APIGatewayV2HTTPResponse{
			Body: gateway.CreateErrorMessage(message, 500), StatusCode: 500}
		return &apiResponse, nil
	}

	jsonBody, _ := json.Marshal(records)
	apiResponse = events.APIGatewayV2HTTPResponse{Body: string(jsonBody), StatusCode: 200}

	return &apiResponse, nil

}
