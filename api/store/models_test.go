package store

import (
	"context"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDBConnection(t *testing.T) {

	// Initiate NEO4j session
	db := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})

	t.Cleanup(func() {
		db.Close(context.Background())
	})

	// Create GraphStore object with initiated db.
	graphStore := NewGraphStore(db)

	err := graphStore.InitOrgAndDataset(1, 1, "N:Org:123", "N:Dataset:123")
	assert.Nil(t, err, "Could not set Org and Dataset in Database")

}

func TestCreateModel(t *testing.T) {
	// Initiate NEO4j session
	db := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})

	t.Cleanup(func() {
		db.Close(context.Background())
	})

	graphStore := NewGraphStore(db)

	model, err := graphStore.CreateModel(1, 1,
		"Model_1", "Model 1", "This is a description", "N:User:1")
	assert.Nil(t, err, "Unable to create model")

	assert.Equal(t, "Model_1", model.Name)
	assert.Equal(t, "Model 1", model.DisplayName)
	assert.Equal(t, "This is a description", model.Description)

}
