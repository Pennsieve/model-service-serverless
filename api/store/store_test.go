package store

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/pennsieve/model-service-serverless/api/models"
	"github.com/pennsieve/model-service-serverless/api/shared"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageState"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageType"
	pgdb2 "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
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
		"create query syntax from query params": testCreateQuery,
		"create org and dataset nodes in db":    testInitOrgAndDataset,
		"create valid model":                    testCreateModel,
		"get package ancestors":                 testPackageAncestors,
	} {
		t.Run(scenario, func(t *testing.T) {
			db := shared.NewNeo4jSession(neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{
				AccessMode: neo4j.AccessModeWrite,
			}))

			pgdbClient, err := pgdb.ConnectENV()
			if err != nil {
				log.Fatal("cannot connect to db:", err)
			}

			graphStore := NewModelServiceStore(pgdbClient, db)

			t.Cleanup(func() {
				db.Close(context.Background())
				pgdbClient.Close()
			})

			fn(t, graphStore)
		})
	}
}

func testCreateQuery(t *testing.T, _ *ModelServiceStore) {

	// Mock result of shortestPath method.
	paths := []dbtype.Path{
		{
			Nodes: []dbtype.Node{
				{
					ElementId: "9609bfb8-c7a1-45d5-b683-de2e39788cc0",
					Props: map[string]any{
						"name": "samples",
						"id":   "9609bfb8-c7a1-45d5-b683-de2e39788cc0",
					},
				},
				{
					ElementId: "42eb7c3e-ac34-4ff1-ad20-672e5b7b97ee",
					Props: map[string]any{
						"name": "visits",
						"id":   "42eb7c3e-ac34-4ff1-ad20-672e5b7b97ee",
					},
				},
				{
					ElementId: "43f44351-7d80-454b-9d11-6ecc0c158559",
					Props: map[string]any{
						"name": "patient",
						"id":   "43f44351-7d80-454b-9d11-6ecc0c158559",
					},
				},
			},
			Relationships: []dbtype.Relationship{
				{
					Props: map[string]any{
						"type": "SAMPLE_BELONGS_TO_VISIT",
					},
				},
				{
					Props: map[string]any{
						"type": "VISIT_BELONGS_TO_SUBJECT",
					},
				},
			},
		},
		{
			Nodes: []dbtype.Node{
				{
					ElementId: "9609bfb8-c7a1-45d5-b683-de2e39788cc0",
					Props: map[string]any{
						"name": "samples",
						"id":   "9609bfb8-c7a1-45d5-b683-de2e39788cc0",
					},
				},
				{
					ElementId: "42eb7c3e-ac34-4ff1-ad20-672e5b7b97ee",
					Props: map[string]any{
						"name": "visits",
						"id":   "42eb7c3e-ac34-4ff1-ad20-672e5b7b97ee",
					},
				},
				{
					ElementId: "0881e84e-7f52-419a-9d9e-6069cae329c5",
					Props: map[string]any{
						"name": "study",
						"id":   "0881e84e-7f52-419a-9d9e-6069cae329c5",
					},
				},
			},
			Relationships: []dbtype.Relationship{
				{
					Props: map[string]any{
						"type": "SAMPLE_BELONGS_TO_VISIT",
					},
				},
				{
					Props: map[string]any{
						"type": "VISIT_BELONGS_TO_STUDY",
					},
				},
			},
		},
		{
			Nodes: []dbtype.Node{
				{
					ElementId: "9609bfb8-c7a1-45d5-b683-de2e39788cc0",
					Props: map[string]any{
						"name": "samples",
						"id":   "9609bfb8-c7a1-45d5-b683-de2e39788cc0",
					},
				},
				{
					ElementId: "42eb7c3e-ac34-4ff1-ad20-672e5b7b97ee",
					Props: map[string]any{
						"name": "visits",
						"id":   "42eb7c3e-ac34-4ff1-ad20-672e5b7b97ee",
					},
				},
				{
					ElementId: "0881e84e-7f52-419a-9d9e-6069cae329c5",
					Props: map[string]any{
						"name": "study",
						"id":   "0881e84e-7f52-419a-9d9e-6069cae329c5",
					},
				},
				{
					ElementId: "0881e84e-7f52-419a-9d9e-6069cae329c5",
					Props: map[string]any{
						"name": "location",
						"id":   "0881e84e-7f52-419a-9d9e-6069cae329c5",
					},
				},
				{
					ElementId: "0881e84e-7f52-419a-9d9e-6069cae329c5",
					Props: map[string]any{
						"name": "state",
						"id":   "0881e84e-7f52-419a-9d9e-6069cae329c5",
					},
				},
			},
			Relationships: []dbtype.Relationship{
				{
					Props: map[string]any{
						"type": "SAMPLE_BELONGS_TO_VISIT",
					},
				},
				{
					Props: map[string]any{
						"type": "VISIT_BELONGS_TO_STUDY",
					},
				},
				{
					Props: map[string]any{
						"type": "STUDY_BELONGS_TO_LOCATION",
					},
				},
				{
					Props: map[string]any{
						"type": "LOCATION_BELONGS_TO_STATE",
					},
				},
			},
		},
	}

	// Example of query filters
	filters := []models.Filters{
		{
			Model:    "patient",
			Property: "name",
			Operator: "STARTS_WITH",
			Value:    "LIM031",
		},
		{
			Model:    "samples",
			Property: "sample_type_id",
			Operator: "STARTS_WITH",
			Value:    "Biopsy Cells",
		},
		{
			Model:    "visit",
			Property: "study",
			Operator: "STARTS_WITH",
			Value:    "Wu LIMBO",
		},
		{
			Model:    "state",
			Property: "mascot",
			Operator: "STARTS_WITH",
			Value:    "Eagle",
		},
	}

	queryStr, err := generateQuery(models.Model{
		ID:   "9609bfb8-c7a1-45d5-b683-de2e39788cc0",
		Name: "samples",
	}, paths, filters, "'@id'", false, "", "")

	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, "MATCH (Msamples:Model{id:'9609bfb8-c7a1-45d5-b683-de2e39788cc0'})<-[:`@INSTANCE_OF`]-(samples:Record)-[:SAMPLE_BELONGS_TO_VISIT]-(visits:Record)-[:VISIT_BELONGS_TO_SUBJECT]-(patient:Record)-[:`@INSTANCE_OF`]->(Mpatient:Model{id:'43f44351-7d80-454b-9d11-6ecc0c158559'}) , (visits:Record)-[:VISIT_BELONGS_TO_STUDY]-(study:Record) , (location:Record)-[:LOCATION_BELONGS_TO_STATE]-(location:Record)-[:LOCATION_BELONGS_TO_STATE]-, (location:Record)-[:LOCATION_BELONGS_TO_STATE]-(state:Record) WHERE patient.name STARTS_WITH 'LIM031' AND samples.sample_type_id STARTS_WITH 'Biopsy Cells' AND visit.study STARTS_WITH 'Wu LIMBO' AND state.mascot STARTS_WITH 'Eagle' RETURN DISTINCT samples AS records ORDER BY samples.'@id' LIMIT 100", queryStr)

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

func testPackageAncestors(t *testing.T, s *ModelServiceStore) {

	orgId := 2
	datasetId := 1

	s.WithOrg(orgId)

	defer func() {
		truncate(t, s.pgdb, orgId, "packages")
		truncate(t, s.pgdb, orgId, "files")
		truncate(t, s.pgdb, orgId, "package_storage")
		truncate(t, s.pgdb, orgId, "organization_storage")
		truncate(t, s.pgdb, orgId, "dataset_storage")
	}()

	// ADD FOLDER TO ROOT
	uploadId, _ := uuid.NewUUID()
	folderParams := pgdb2.PackageParams{
		Name:         "Folder1",
		PackageType:  packageType.Collection,
		PackageState: packageState.Ready,
		NodeId:       fmt.Sprintf("N:Package:%s", uploadId.String()),
		ParentId:     -1,
		DatasetId:    datasetId,
		OwnerId:      1,
		Size:         1000, // should be ignored
		ImportId:     sql.NullString{String: uploadId.String(), Valid: true},
		Attributes:   []packageInfo.PackageAttribute{},
	}

	folder1, err := s.pg.AddFolder(context.Background(), folderParams)
	assert.NoError(t, err)

	// ADD NESTED FOLDER
	uploadId, _ = uuid.NewUUID()
	folderParams = pgdb2.PackageParams{
		Name:         "Folder2",
		PackageType:  packageType.Collection,
		PackageState: packageState.Ready,
		NodeId:       fmt.Sprintf("N:Package:%s", uploadId.String()),
		ParentId:     folder1.Id,
		DatasetId:    datasetId,
		OwnerId:      1,
		Size:         1000, // should be ignored
		ImportId:     sql.NullString{String: uploadId.String(), Valid: true},
		Attributes:   []packageInfo.PackageAttribute{},
	}

	folder2, err := s.pg.AddFolder(context.Background(), folderParams)
	assert.NoError(t, err)

	// Test adding packages to root
	testPackageNodeId := "N:Package:1"
	testParams := []testPackageParams{
		{Name: "package_7.txt", ParentId: folder2.Id, NodeId: testPackageNodeId},
	}

	insertParams := GenerateTestPackages(testParams, datasetId)
	_, err = s.pg.AddPackages(context.Background(), insertParams)
	assert.NoError(t, err)

	ctx := context.Background()
	ancestors, err := s.pg.GetPackageAncestors(ctx, testPackageNodeId)
	assert.NoError(t, err)
	assert.Len(t, ancestors, 3)
	assert.Equal(t, testPackageNodeId, ancestors[0].NodeId, "Expecting the first index of result to be the requested package")
	assert.Equal(t, folderParams.NodeId, ancestors[1].NodeId, "Expecting 2nd index of result to be the nested folder")
	assert.Equal(t, false, ancestors[2].ParentId.Valid, "Expecting 3rd index of result to be the root folder")

}

type testPackageParams struct {
	Name     string
	ParentId int64
	NodeId   string
}

func truncate(t *testing.T, db *sql.DB, orgID int, table string) {

	var query string

	switch table {
	case "organization_storage":
		query = fmt.Sprintf("TRUNCATE TABLE pennsieve.%s CASCADE", table)
	default:
		query = fmt.Sprintf("TRUNCATE TABLE \"%d\".%s CASCADE", orgID, table)
	}

	_, err := db.Exec(query)
	assert.NoError(t, err)
}

func GenerateTestPackages(params []testPackageParams, datasetId int) []pgdb2.PackageParams {

	var result []pgdb2.PackageParams

	attr := []packageInfo.PackageAttribute{
		{
			Key:      "subtype",
			Fixed:    false,
			Value:    "Image",
			Hidden:   true,
			Category: "Pennsieve",
			DataType: "string",
		}, {
			Key:      "icon",
			Fixed:    false,
			Value:    "Microscope",
			Hidden:   true,
			Category: "Pennsieve",
			DataType: "string",
		},
	}

	for _, p := range params {
		var uploadId string
		var nodeId string
		if p.NodeId == "" {
			u, _ := uuid.NewUUID()
			uploadId = u.String()
			nodeId = fmt.Sprintf("N:Package:%s", u.String())
		} else {
			u, _ := uuid.NewUUID()
			uploadId = u.String()
			nodeId = p.NodeId
		}

		insertPackage := pgdb2.PackageParams{
			Name:         p.Name,
			PackageType:  packageType.Image,
			PackageState: packageState.Unavailable,
			NodeId:       nodeId,
			ParentId:     p.ParentId,
			DatasetId:    datasetId,
			OwnerId:      1,
			Size:         1000,
			ImportId: sql.NullString{
				String: uploadId,
				Valid:  true,
			},
			Attributes: attr,
		}

		result = append(result, insertPackage)
	}

	return result
}
