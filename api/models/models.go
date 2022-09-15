package models

import (
	"context"
	"fmt"
	"github.com/pennsieve/model-service-serverless/api/core"
	"log"
	"strings"
	"time"
)

type ModelDTO struct {
	Count         int64       `json:"count"`
	CreatedAt     time.Time   `json:"createdAt"`
	CreatedBy     string      `json:"createdBy"`
	Description   string      `json:"description"`
	DisplayName   string      `json:"displayName"`
	ID            string      `json:"id"`
	Locked        bool        `json:"locked"`
	Name          string      `json:"name"`
	PropertyCount int64       `json:"propertyCount"`
	TemplateID    interface{} `json:"templateId"`
	UpdatedAt     time.Time   `json:"updatedAt"`
	UpdatedBy     string      `json:"updatedBy"`
}

// GetModels returns a list of models for a provided dataset within an organization
func GetModels(session core.Neo4jAPI, datasetId int, organizationId int) (*[]ModelDTO, error) {

	var cql strings.Builder
	cql.WriteString("MATCH  (m:Model)")
	cql.WriteString(fmt.Sprintf("-[:`@IN_DATASET`]->(:Dataset { id: %d }) ", datasetId))
	cql.WriteString(fmt.Sprintf("-[:`@IN_ORGANIZATION`]->(:Organization { id: %d }) ", organizationId))
	cql.WriteString("MATCH (m)-[created:`@CREATED_BY`]->(c:User)")
	cql.WriteString("MATCH (m)-[updated:`@UPDATED_BY`]->(u:User)")
	cql.WriteString("OPTIONAL MATCH (m)-[r:`@RELATED_TO`]->(n) WHERE r.index IS NOT NULL ")
	cql.WriteString("RETURN m.name AS name, m.description AS description, m.id AS id, m.display_name AS display_name,")
	cql.WriteString("size(()-[:`@INSTANCE_OF`]->(m)) AS count,")
	cql.WriteString("size((m)-[:`@HAS_PROPERTY`]->()) AS nrStaticProps, count((m)--(n)) AS nrLinkedProps,")
	cql.WriteString("c.node_id AS created_by, u.node_id AS updated_by, created.at AS created_at, updated.at AS updated_at")

	ctx := context.Background()
	transaction, err := session.BeginTransaction(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println(cql.String())
	result, err := transaction.Run(ctx, cql.String(), nil)
	if err != nil {
		return nil, err
	}

	// Iterate over results and create array of model objects
	var models []ModelDTO
	for result.Next(ctx) {
		log.Println("interation")
		record := result.Record()

		// Create key/value map based on keys and values returned.
		valueMap := make(map[string]interface{})
		values := record.Values
		for i, k := range record.Keys {
			valueMap[k] = values[i]
		}

		log.Println(values)

		m := ModelDTO{
			Count:         valueMap["count"].(int64),
			CreatedAt:     time.Now(), //valueMap["created_at"].(time.Time),
			CreatedBy:     stringOrEmpty(valueMap["created_by"]),
			Description:   stringOrEmpty(valueMap["description"]),
			DisplayName:   stringOrEmpty(valueMap["display_name"]),
			ID:            stringOrEmpty(valueMap["id"]),
			Locked:        false,
			Name:          stringOrEmpty(valueMap["name"]),
			PropertyCount: valueMap["nrStaticProps"].(int64) + valueMap["nrLinkedProps"].(int64),
			TemplateID:    nil,
			UpdatedAt:     time.Now(), //valueMap["updated_at"].(time.Time),
			UpdatedBy:     stringOrEmpty(valueMap["updated_by"]),
		}

		models = append(models, m)
	}
	// Err returns the error that caused Next to return false
	if err = result.Err(); err != nil {
		return nil, err
	}

	transaction.Close(ctx)

	return &models, nil

}

func stringOrEmpty(v interface{}) string {
	if v != nil {
		return v.(string)
	}
	return ""
}
