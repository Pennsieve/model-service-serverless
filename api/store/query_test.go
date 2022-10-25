package store

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/pennsieve/model-service-serverless/api/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQuery(t *testing.T) {

	paths := []dbtype.Path{
		dbtype.Path{
			Nodes: []dbtype.Node{
				dbtype.Node{
					Id:        0,
					ElementId: "9609bfb8-c7a1-45d5-b683-de2e39788cc0",
					Labels:    nil,
					Props: map[string]any{
						"name": "samples",
						"id":   "9609bfb8-c7a1-45d5-b683-de2e39788cc0",
					},
				},
				dbtype.Node{
					Id:        1,
					ElementId: "42eb7c3e-ac34-4ff1-ad20-672e5b7b97ee",
					Labels:    nil,
					Props: map[string]any{
						"name": "visits",
						"id":   "42eb7c3e-ac34-4ff1-ad20-672e5b7b97ee",
					},
				},
				dbtype.Node{
					Id:        2,
					ElementId: "43f44351-7d80-454b-9d11-6ecc0c158559",
					Labels:    nil,
					Props: map[string]any{
						"name": "patient",
						"id":   "43f44351-7d80-454b-9d11-6ecc0c158559",
					},
				},
			},
			Relationships: []dbtype.Relationship{
				dbtype.Relationship{
					Id:             0,
					ElementId:      "",
					StartId:        0,
					StartElementId: "",
					EndId:          0,
					EndElementId:   "",
					Type:           "",
					Props: map[string]any{
						"type": "SAMPLE_BELONGS_TO_VISIT",
					},
				},
				dbtype.Relationship{
					Id:             1,
					ElementId:      "",
					StartId:        0,
					StartElementId: "",
					EndId:          0,
					EndElementId:   "",
					Type:           "",
					Props: map[string]any{
						"type": "VISIT_BELONGS_TO_SUBJECT",
					},
				},
			},
		},
		dbtype.Path{
			Nodes: []dbtype.Node{
				dbtype.Node{
					Id:        0,
					ElementId: "9609bfb8-c7a1-45d5-b683-de2e39788cc0",
					Labels:    nil,
					Props: map[string]any{
						"name": "samples",
						"id":   "9609bfb8-c7a1-45d5-b683-de2e39788cc0",
					},
				},
				dbtype.Node{
					Id:        1,
					ElementId: "42eb7c3e-ac34-4ff1-ad20-672e5b7b97ee",
					Labels:    nil,
					Props: map[string]any{
						"name": "visits",
						"id":   "42eb7c3e-ac34-4ff1-ad20-672e5b7b97ee",
					},
				},
				dbtype.Node{
					Id:        2,
					ElementId: "0881e84e-7f52-419a-9d9e-6069cae329c5",
					Labels:    nil,
					Props: map[string]any{
						"name": "study",
						"id":   "0881e84e-7f52-419a-9d9e-6069cae329c5",
					},
				},
			},
			Relationships: []dbtype.Relationship{
				dbtype.Relationship{
					Id:             0,
					ElementId:      "",
					StartId:        0,
					StartElementId: "",
					EndId:          0,
					EndElementId:   "",
					Type:           "",
					Props: map[string]any{
						"type": "SAMPLE_BELONGS_TO_VISIT",
					},
				},
				dbtype.Relationship{
					Id:             2,
					ElementId:      "",
					StartId:        0,
					StartElementId: "",
					EndId:          0,
					EndElementId:   "",
					Type:           "",
					Props: map[string]any{
						"type": "VISIT_BELONGS_TO_STUDY",
					},
				},
			},
		},
		dbtype.Path{
			Nodes: []dbtype.Node{
				dbtype.Node{
					Id:        0,
					ElementId: "9609bfb8-c7a1-45d5-b683-de2e39788cc0",
					Labels:    nil,
					Props: map[string]any{
						"name": "samples",
						"id":   "9609bfb8-c7a1-45d5-b683-de2e39788cc0",
					},
				},
				dbtype.Node{
					Id:        1,
					ElementId: "42eb7c3e-ac34-4ff1-ad20-672e5b7b97ee",
					Labels:    nil,
					Props: map[string]any{
						"name": "visits",
						"id":   "42eb7c3e-ac34-4ff1-ad20-672e5b7b97ee",
					},
				},
				dbtype.Node{
					Id:        2,
					ElementId: "0881e84e-7f52-419a-9d9e-6069cae329c5",
					Labels:    nil,
					Props: map[string]any{
						"name": "study",
						"id":   "0881e84e-7f52-419a-9d9e-6069cae329c5",
					},
				},
				dbtype.Node{
					Id:        2,
					ElementId: "0881e84e-7f52-419a-9d9e-6069cae329c5",
					Labels:    nil,
					Props: map[string]any{
						"name": "location",
						"id":   "0881e84e-7f52-419a-9d9e-6069cae329c5",
					},
				},
				dbtype.Node{
					Id:        3,
					ElementId: "0881e84e-7f52-419a-9d9e-6069cae329c5",
					Labels:    nil,
					Props: map[string]any{
						"name": "state",
						"id":   "0881e84e-7f52-419a-9d9e-6069cae329c5",
					},
				},
			},
			Relationships: []dbtype.Relationship{
				dbtype.Relationship{
					Id:             0,
					ElementId:      "",
					StartId:        0,
					StartElementId: "",
					EndId:          0,
					EndElementId:   "",
					Type:           "",
					Props: map[string]any{
						"type": "SAMPLE_BELONGS_TO_VISIT",
					},
				},
				dbtype.Relationship{
					Id:             2,
					ElementId:      "",
					StartId:        0,
					StartElementId: "",
					EndId:          0,
					EndElementId:   "",
					Type:           "",
					Props: map[string]any{
						"type": "VISIT_BELONGS_TO_STUDY",
					},
				},
				dbtype.Relationship{
					Id:             3,
					ElementId:      "",
					StartId:        0,
					StartElementId: "",
					EndId:          0,
					EndElementId:   "",
					Type:           "",
					Props: map[string]any{
						"type": "STUDY_BELONGS_TO_LOCATION",
					},
				},
				dbtype.Relationship{
					Id:             4,
					ElementId:      "",
					StartId:        0,
					StartElementId: "",
					EndId:          0,
					EndElementId:   "",
					Type:           "",
					Props: map[string]any{
						"type": "LOCATION_BELONGS_TO_STATE",
					},
				},
			},
		},
	}

	filters := []models.Filters{
		models.Filters{
			Model:    "patient",
			Property: "name",
			Operator: "STARTS_WITH",
			Value:    "LIM031",
		},
		models.Filters{
			Model:    "samples",
			Property: "sample_type_id",
			Operator: "STARTS_WITH",
			Value:    "Biopsy Cells",
		},
		models.Filters{
			Model:    "visit",
			Property: "study",
			Operator: "STARTS_WITH",
			Value:    "Wu LIMBO",
		},
		models.Filters{
			Model:    "state",
			Property: "mascot",
			Operator: "STARTS_WITH",
			Value:    "Eagle",
		},
	}

	queryStr, _ := generateQuery(paths, filters)

	assert.Equal(t, "MATCH (Msamples:Model{id:'9609bfb8-c7a1-45d5-b683-de2e39788cc0'})<-[:`@INSTANCE_OF`]-(samples:Record)-[:SAMPLE_BELONGS_TO_VISIT]-(visits:Record)-[:VISIT_BELONGS_TO_SUBJECT]-(patient:Record)-[:`@INSTANCE_OF`]->(Mpatient:Model{id:'43f44351-7d80-454b-9d11-6ecc0c158559'}) , (visits:Record)-[:VISIT_BELONGS_TO_STUDY]-(study:Record) , (location:Record)-[:LOCATION_BELONGS_TO_STATE]-(location:Record)-[:LOCATION_BELONGS_TO_STATE]-, (location:Record)-[:LOCATION_BELONGS_TO_STATE]-(state:Record) WHERE patient.name STARTS_WITH 'LIM031' AND samples.sample_type_id STARTS_WITH 'Biopsy Cells' AND visit.study STARTS_WITH 'Wu LIMBO' AND state.mascot STARTS_WITH 'Eagle' RETURN samples AS records LIMIT 100", queryStr)

	fmt.Println(queryStr)

}
