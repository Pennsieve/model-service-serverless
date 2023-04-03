package store

import (
	"context"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/pennsieve/model-service-serverless/api/models"
	"github.com/pennsieve/model-service-serverless/api/shared"
	log "github.com/sirupsen/logrus"
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
		time.Sleep(30 * time.Second)
	}

	// Get Connection
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

func TestStore(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, s *ModelServiceStore,
	){
		"create org and dataset nodes in db": testInitOrgAndDataset,
		"create valid model":                 testCreateModel,
	} {
		t.Run(scenario, func(t *testing.T) {
			db := shared.NewNeo4jSession(neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{
				AccessMode: neo4j.AccessModeWrite,
			}))

			graphStore := NewModelServiceStore(db)

			t.Cleanup(func() {
				db.Close(context.Background())
			})

			fn(t, graphStore)
		})
	}
}

func testInitOrgAndDataset(t *testing.T, s *ModelServiceStore) {
	ctx := context.Background()
	err := s.neo.InitOrgAndDataset(ctx, 1, 1, "N:Org:123", "N:Dataset:123")
	assert.Nil(t, err, "Could not set Org and Dataset in Database")

}

func testCreateModel(t *testing.T, s *ModelServiceStore) {
	// Initiate NEO4j session

	// Create Model in database
	t.Run("create model", func(t *testing.T) {
		ctx := context.Background()
		model, err := s.neo.CreateModel(ctx, 1, 1,
			"Model_1", "Model 1", "This is a description", "N:User:1")
		assert.Nil(t, err, "Unable to create model")
		assert.Equal(t, "Model_1", model.Name)
		assert.Equal(t, "Model 1", model.DisplayName)
		assert.Equal(t, "This is a description", model.Description)
	})

	t.Run("error on model with wxisting name", func(t *testing.T) {
		ctx := context.Background()
		_, err := s.neo.CreateModel(ctx, 1, 1,
			"Model_1", "Model 1", "This is a description", "N:User:1")
		if assert.Error(t, err) {
			assert.Equal(t, &models.ModelNameCountError{Name: "Model_1"}, err)
		}
	})

	t.Cleanup(func() {
		cql := "MATCH (n:Model {name: 'Model_1'}) OPTIONAL MATCH (n)-[r]-() DELETE n, r"
		_, err := s.neodb.Run(context.Background(), cql, nil)
		if err != nil {
			log.Fatalln(err)
		}
	})

}
