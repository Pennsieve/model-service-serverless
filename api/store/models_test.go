package store

import (
	"context"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/pennsieve/model-service-serverless/api/models"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestStore(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, s *graphStore,
	){
		"create org and dataset nodes in db": testInitOrgAndDataset,
		"create valid model":                 testCreateModel,
	} {
		t.Run(scenario, func(t *testing.T) {
			db := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{
				AccessMode: neo4j.AccessModeWrite,
			})

			graphStore := NewGraphStore(db)

			t.Cleanup(func() {
				db.Close(context.Background())
			})

			fn(t, graphStore)
		})
	}
}

func testInitOrgAndDataset(t *testing.T, s *graphStore) {
	err := s.InitOrgAndDataset(1, 1, "N:Org:123", "N:Dataset:123")
	assert.Nil(t, err, "Could not set Org and Dataset in Database")

}

func testCreateModel(t *testing.T, s *graphStore) {
	// Initiate NEO4j session

	// Create Model in database
	t.Run("create model", func(t *testing.T) {
		model, err := s.CreateModel(1, 1,
			"Model_1", "Model 1", "This is a description", "N:User:1")
		assert.Nil(t, err, "Unable to create model")
		assert.Equal(t, "Model_1", model.Name)
		assert.Equal(t, "Model 1", model.DisplayName)
		assert.Equal(t, "This is a description", model.Description)
	})

	t.Run("error on model with wxisting name", func(t *testing.T) {
		_, err := s.CreateModel(1, 1,
			"Model_1", "Model 1", "This is a description", "N:User:1")
		if assert.Error(t, err) {
			assert.Equal(t, &models.ModelNameCountError{Name: "Model_1"}, err)
		}
	})

	t.Cleanup(func() {
		cql := "MATCH (n:Model {name: 'Model_1'}) OPTIONAL MATCH (n)-[r]-() DELETE n, r"
		_, err := s.db.Run(context.Background(), cql, nil)
		if err != nil {
			log.Fatalln(err)
		}
	})

}
