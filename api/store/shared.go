package store

import (
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/db"
	"github.com/pennsieve/model-service-serverless/api/models"
	"time"
)

// set Temporary set used to check if string in list of other strings.
type set map[string]struct{}

// has checks if the map contains a provided string
func (s set) has(v string) bool {
	_, ok := s[v]
	return ok
}

// newSet local method to turn list of Records into map on provided key.
func newSet(slice []*db.Record, p string) set {
	s := make(set)
	for _, each := range slice {
		v, _ := each.Get(p)
		s[v.(string)] = struct{}{}
	}
	return s
}

// mapNodes maps the origin nodes and target nodes into map for neo4j query.
func mapNodes(originNodes []string, targetNodes []string) []map[string]interface{} {
	var result = make([]map[string]interface{}, len(originNodes))

	for index, _ := range targetNodes {
		result[index] = map[string]interface{}{"from": originNodes[index], "to": targetNodes[index], "uuid": uuid.New().String()}
	}

	return result
}

// parseRecordRelationshipResponse returns a Model object parsed from a Neo4J query result.
func parseRecordRelationshipResponse(record *db.Record, relType string) models.ShortRecordRelationShip {
	// Create key/value map based on keys and values returned.
	valueMap := make(map[string]interface{})
	values := record.Values
	for i, k := range record.Keys {
		valueMap[k] = values[i]
	}

	m := models.ShortRecordRelationShip{
		ID:      stringOrEmpty(valueMap["relID"]),
		From:    stringOrEmpty(valueMap["from"]),
		To:      stringOrEmpty(valueMap["to"]),
		RelType: relType,
	}

	return m
}

func stringOrEmpty(v interface{}) string {
	if v != nil {
		return v.(string)
	}
	return ""
}

// parseResponse returns a Model object parsed from a Neo4J query result.
func parseModelResponse(record *db.Record) models.Model {
	// Create key/value map based on keys and values returned.
	valueMap := make(map[string]interface{})
	values := record.Values
	for i, k := range record.Keys {
		valueMap[k] = values[i]
	}

	m := models.Model{
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
	return m
}

func parseModelPropertyResponse(record *db.Record) models.ModelProperty {
	// Create key/value map based on keys and values returned.
	valueMap := make(map[string]interface{})
	values := record.Values
	for i, k := range record.Keys {
		valueMap[k] = values[i]
	}

	p := models.ModelProperty{
		ID:           stringOrEmpty(valueMap["id"]),
		DataType:     valueMap["data_type"],
		DefaultValue: valueMap["default_value"],
		DisplayName:  stringOrEmpty(valueMap["display_name"]),
		Name:         stringOrEmpty(valueMap["name"]),
		IsModelTitle: valueMap["model_title"].(bool),
		Index:        valueMap["index"].(int64),
	}
	return p
}
