package service

import (
	"github.com/pennsieve/model-service-serverless/api/models"
	"github.com/pennsieve/model-service-serverless/api/store"
)

type GraphService interface {
	GetDatasetModels(datasetId int, organizationId int) ([]models.Model, error)
	QueryGraph(parsedRequestBody models.QueryRequestBody, datasetId int,
		organizationId int) ([]models.Record, error)
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

	var models []models.Model
	for _, v := range results {
		models = append(models, v)
	}

	return models, nil
}

func (s *graphService) QueryGraph(parsedRequestBody models.QueryRequestBody, datasetId int,
	organizationId int) ([]models.Record, error) {

	nodes, err := s.store.Query(datasetId, organizationId, parsedRequestBody)

	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (s *graphService) CreateRelationships(parsedRequestBody models.PostRecordRelationshipRequestBody,
	datasetId int, organizationId int, userNodeId string) ([]models.ShortRecordRelationShip, error) {

	response, err := s.store.CreateRelationShips(datasetId, organizationId, userNodeId, parsedRequestBody)
	if err != nil {
		return nil, err
	}

	return response, nil
}
