package store

import (
	"context"
	"fmt"
	"github.com/pennsieve/model-service-serverless/api/models"
	"github.com/pennsieve/model-service-serverless/api/queries"
	"github.com/pennsieve/model-service-serverless/api/shared"
)

// ModelServiceStore provides the Queries interface and a db instance.
type ModelServiceStore struct {
	neo   *queries.NeoQueries
	neodb *shared.Neo4jSession
}

// NewModelServiceStore returns a UploadHandlerStore object which implements the Queries
func NewModelServiceStore(neo *shared.Neo4jSession) *ModelServiceStore {
	return &ModelServiceStore{
		neo:   queries.NewNeoQueries(neo),
		neodb: neo,
	}
}

func (s *ModelServiceStore) execTx(ctx context.Context, fn func(queries *queries.NeoQueries) error) error {

	// NOTE: When you create a new transaction (as below), the s.pgdb is NOT part of the transaction.
	// This has the following impact
	// 1. If you have set the search-path for the pgdb, the search path is no longer applied to s.pgdb
	// 2. Any function that is wrapped in the execTx method should ONLY use the provided queries' struct that wraps the transaction.
	// 3. To enable custom Queries for a service, we wrap the pgdb.Queries in a service specific Queries struct.
	//	  This enables you to create custom queries within the service that leverage the transaction
	//    You can use the exposed db property of the Queries' struct to create custom database interactions.
	//	  See the "upload-service-v2/upload lambda" for an example

	tx, err := s.neodb.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	q := queries.NewNeoQueries(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

func (s *ModelServiceStore) CreateModelTx(ctx context.Context, datasetId int, organizationId int, name string, displayName string, description string, userId string) (*models.Model, error) {

	var createdModel *models.Model
	err := s.execTx(ctx, func(qtx *queries.NeoQueries) error {
		var err error
		createdModel, err = qtx.CreateModel(ctx, datasetId, organizationId, name, displayName, description, userId)
		return err
	})
	if err != nil {
		return nil, err
	}

	return createdModel, err

}

func (s *ModelServiceStore) GetDatasetModels(ctx context.Context, datasetId int, organizationId int) ([]models.Model, error) {
	// Get the models from Neo4J
	results, err := s.neo.GetModels(ctx, datasetId, organizationId)
	if err != nil {
		return nil, err
	}

	var models []models.Model
	for _, v := range results {
		models = append(models, v)
	}

	return models, nil
}

func (s *ModelServiceStore) QueryGraph(ctx context.Context, parsedRequestBody models.QueryRequestBody, datasetId int,
	organizationId int) ([]models.Record, error) {

	nodes, err := s.neo.Query(ctx, datasetId, organizationId, parsedRequestBody)

	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (s *ModelServiceStore) Autocomplete(ctx context.Context, parsedRequestBody models.AutocompleteRequestBody, datasetId int,
	organizationId int) ([]string, error) {

	values, err := s.neo.Autocomplete(ctx, datasetId, organizationId, parsedRequestBody)

	if err != nil {
		return nil, err
	}

	return values, nil
}

func (s *ModelServiceStore) CreateRelationships(ctx context.Context, parsedRequestBody models.PostRecordRelationshipRequestBody,
	datasetId int, organizationId int, userNodeId string) ([]models.ShortRecordRelationShip, error) {

	var response []models.ShortRecordRelationShip
	err := s.execTx(ctx, func(qtx *queries.NeoQueries) error {
		var err error
		response, err = s.neo.CreateRelationShips(ctx, datasetId, organizationId, userNodeId, parsedRequestBody)
		return err
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ModelServiceStore) GetRecordsForPackage(ctx context.Context, datasetId int, organizationId int, packageNodeId string, maxDepth int) ([]models.Record, error) {

	nodes, err := s.neo.GetRecordsForPackage(ctx, datasetId, organizationId, packageNodeId, maxDepth)

	if err != nil {
		return nil, err
	}

	return nodes, nil

}
