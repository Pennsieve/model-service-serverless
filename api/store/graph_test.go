package store

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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
