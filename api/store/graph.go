package store

import (
	"context"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/pennsieve/model-service-serverless/api/models"
)

type GraphStore interface {
	CreateRelationShips(datasetId int, organizationId int, userId string,
		q models.PostRecordRelationshipRequestBody) ([]models.ShortRecordRelationShip, error)
	GetModelByName(modelName string, datasetId int, organizationId int) (*models.Model, error)
	GetModels(datasetId int, organizationId int) (map[string]models.Model, error)
	Query(datasetId int, organizationId int, q models.QueryRequestBody) ([]models.Record, error)
	Autocomplete(datasetId int, organizationId int, q models.AutocompleteRequestBody) ([]string, error)
	ShortestPath(ctx context.Context, sourceModel models.Model, targetModels map[string]string) ([]dbtype.Path, error)
	CreateModel(datasetId int, organizationId int, name string, displayName string, description string, userId string) (*models.Model, error)
	InitOrgAndDataset(organizationId int, datasetId int, organizationNodeId string, datasetNodeId string) error
	GetRecordsForPackage(ctx context.Context, datasetId int, organizationId int, packageNodeId string, maxDepth int) ([]models.Record, error)
}

func NewGraphStore(db neo4j.SessionWithContext) *graphStore {
	return &graphStore{
		db: db,
	}
}

type graphStore struct {
	db neo4j.SessionWithContext
}
