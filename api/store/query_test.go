package store

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/pennsieve/model-service-serverless/api/models"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestFail(t *testing.T) {
	log.Println("Testing")
	assert.Equal(t, true, true)
}

func TestQuery(t *testing.T) {

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
