package store

import (
	"context"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/pennsieve/model-service-serverless/api/models"
)

type Neo4jAPI interface {
	BeginTransaction(ctx context.Context, configurers ...func(*neo4j.TransactionConfig)) (neo4j.ExplicitTransaction, error)
	ExecuteRead(ctx context.Context, work neo4j.ManagedTransactionWork, configurers ...func(*neo4j.TransactionConfig)) (any, error)
	ExecuteWrite(ctx context.Context, work neo4j.ManagedTransactionWork, configurers ...func(*neo4j.TransactionConfig)) (any, error)
	Run(ctx context.Context, cypher string, params map[string]any, configurers ...func(*neo4j.TransactionConfig)) (neo4j.ResultWithContext, error)
	Close(ctx context.Context) error
}

type GraphStore interface {
	CreateRelationShips(datasetId int, organizationId int, userId string,
		q models.PostRecordRelationshipRequestBody) ([]models.ShortRecordRelationShip, error)
	GetModelByName(modelName string, datasetId int, organizationId int) (*models.Model, error)
	GetModels(datasetId int, organizationId int) ([]models.Model, error)
	Query(datasetId int, organizationId int, q models.QueryRequestBody) error
}

func NewGraphStore(db Neo4jAPI) *graphStore {
	return &graphStore{
		db: db,
	}
}

type graphStore struct {
	db Neo4jAPI
}
