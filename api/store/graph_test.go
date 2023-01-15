package store

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"os"
	"testing"
	"time"
)

var neo4jDriver neo4j.DriverWithContext

func TestMain(m *testing.M) {

	// Wait a couple of seconds to enable NEO4J to start up.
	time.Sleep(5 * time.Second)

	// Get Connection
	testDBUri := "bolt://db:7687"
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
