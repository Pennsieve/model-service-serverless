package service

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/pennsieve/model-service-serverless/api/models"
	"github.com/pennsieve/model-service-serverless/api/store"
)

type GraphService interface {
	GetDatasetModels(datasetId int, organizationId int) ([]models.Model, error)
	QueryGraph(parsedRequestBody models.QueryRequestBody, datasetId int,
		organizationId int) (*events.APIGatewayV2HTTPResponse, error)
	CreateRelationships(parsedRequestBody models.PostRecordRelationshipRequestBody,
		datasetId int, organizationId int, userNodeId string) ([]models.ShortRecordRelationShip, error)
}

func NewGraphService(store store.GraphStore) *graphService {
	return &graphService{
		store: store,
	}
}

type graphService struct {
	store store.GraphStore
}

func (s *graphService) GetDatasetModels(datasetId int, organizationId int) ([]models.Model, error) {
	// Get the models from Neo4J
	results, err := s.store.GetModels(datasetId, organizationId)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *graphService) QueryGraph(parsedRequestBody models.QueryRequestBody, datasetId int,
	organizationId int) (*events.APIGatewayV2HTTPResponse, error) {
	apiResponse := events.APIGatewayV2HTTPResponse{}

	err := s.store.Query(datasetId, organizationId, parsedRequestBody)

	if err != nil {
		return nil, err
	}

	// CREATING API RESPONSE
	responseBody := "Success"
	jsonBody, _ := json.Marshal(responseBody)
	apiResponse = events.APIGatewayV2HTTPResponse{Body: string(jsonBody), StatusCode: 200}

	return &apiResponse, nil
}

func (s *graphService) CreateRelationships(parsedRequestBody models.PostRecordRelationshipRequestBody,
	datasetId int, organizationId int, userNodeId string) ([]models.ShortRecordRelationShip, error) {

	response, err := s.store.CreateRelationShips(datasetId, organizationId, userNodeId, parsedRequestBody)
	if err != nil {
		return nil, err
	}

	return response, nil
}
