package service

import (
	"context"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/pennsieve/model-service-serverless/api/store"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var neo4jDriver neo4j.DriverWithContext

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func TestMain(m *testing.M) {

	// If testing on Jenkins (-> NEO4J_BOLT_URL is set) then wait for db to be active.
	if _, ok := os.LookupEnv("NEO4J_BOLT_URL"); ok {
		time.Sleep(120 * time.Second)
	}

	// Get NEO4J URI based on local vs. dockerized testing
	testDBUri := getEnv("NEO4J_BOLT_URL", "bolt://localhost:7687")
	testUserName := "neo4j"
	testPassword := "blackandwhite"

	var err error
	neo4jDriver, err = neo4j.NewDriverWithContext(testDBUri,
		neo4j.BasicAuth(testUserName, testPassword, ""),
		func(config *neo4j.Config) {
			config.MaxConnectionPoolSize = 10
			config.MaxConnectionLifetime = 5 * time.Minute
			config.ConnectionAcquisitionTimeout = 10 * time.Second
		})
	if err != nil {
		panic(err)
	}

	// Seed NEO4J database

	// Run tests
	code := m.Run()

	// return
	os.Exit(code)
}

func TestDBConnection(t *testing.T) {

	// Initiate NEO4j session
	db := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})

	t.Cleanup(func() {
		db.Close(context.Background())
	})

	// Create GraphStore object with initiated db.
	graphStore := store.NewGraphStore(db)
	service := NewGraphService(graphStore)

	err := service.store.InitOrgAndDataset(2, 1, "N:Organization:123", "N:Dataset:123")
	assert.Nil(t, err, "Could not set organization and dataset")

	models, err := service.GetDatasetModels(1, 2)
	assert.Nil(t, err, "Could not get Dataset Models")

	assert.Equal(t, 0, len(models))

}
