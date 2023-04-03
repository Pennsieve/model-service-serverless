package shared

import (
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/db"
	"github.com/pennsieve/model-service-serverless/api/models"
	"time"
)

// set Temporary set used to check if string in list of other strings.
type set map[string]struct{}

// Has checks if the map contains a provided string
func (s set) Has(v string) bool {
	_, ok := s[v]
	return ok
}

// NewSet local method to turn list of Records into map on provided key.
func NewSet(slice []*db.Record, p string) set {
	s := make(set)
	for _, each := range slice {
		v, _ := each.Get(p)
		s[v.(string)] = struct{}{}
	}
	return s
}

// MapNodes maps the origin nodes and target nodes into map for neo4j query.
func MapNodes(originNodes []string, targetNodes []string) []map[string]interface{} {
	var result = make([]map[string]interface{}, len(originNodes))

	for index, _ := range targetNodes {
		result[index] = map[string]interface{}{"from": originNodes[index], "to": targetNodes[index], "uuid": uuid.New().String()}
	}

	return result
}

// ParseRecordRelationshipResponse returns a Model object parsed from a Neo4J query result.
func ParseRecordRelationshipResponse(record *db.Record, relType string) models.ShortRecordRelationShip {
	// Create key/value map based on keys and values returned.
	valueMap := make(map[string]interface{})
	values := record.Values
	for i, k := range record.Keys {
		valueMap[k] = values[i]
	}

	m := models.ShortRecordRelationShip{
		ID:      StringOrEmpty(valueMap["relID"]),
		From:    StringOrEmpty(valueMap["from"]),
		To:      StringOrEmpty(valueMap["to"]),
		RelType: relType,
	}

	return m
}

func StringOrEmpty(v interface{}) string {
	if v != nil {
		return v.(string)
	}
	return ""
}

// StringInSlice checks if a string exists in an array of strings
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func ParseModelPropertyResponse(record *db.Record) models.ModelProperty {
	// Create key/value map based on keys and values returned.
	valueMap := make(map[string]interface{})
	values := record.Values
	for i, k := range record.Keys {
		valueMap[k] = values[i]
	}

	p := models.ModelProperty{
		ID:           StringOrEmpty(valueMap["id"]),
		DataType:     valueMap["data_type"],
		DefaultValue: valueMap["default_value"],
		DisplayName:  StringOrEmpty(valueMap["display_name"]),
		Name:         StringOrEmpty(valueMap["name"]),
		IsModelTitle: valueMap["model_title"].(bool),
		Index:        valueMap["index"].(int64),
	}
	return p
}

// ParseModelResponse returns a Model object parsed from a Neo4J query result.
func ParseModelResponse(record *db.Record) models.Model {
	// Create key/value map based on keys and values returned.
	valueMap := make(map[string]interface{})
	values := record.Values
	for i, k := range record.Keys {
		valueMap[k] = values[i]
	}

	m := models.Model{
		Count:         valueMap["count"].(int64),
		CreatedAt:     time.Now(), //valueMap["created_at"].(time.Time),
		CreatedBy:     StringOrEmpty(valueMap["created_by"]),
		Description:   StringOrEmpty(valueMap["description"]),
		DisplayName:   StringOrEmpty(valueMap["display_name"]),
		ID:            StringOrEmpty(valueMap["id"]),
		Locked:        false,
		Name:          StringOrEmpty(valueMap["name"]),
		PropertyCount: valueMap["nrStaticProps"].(int64) + valueMap["nrLinkedProps"].(int64),
		TemplateID:    nil,
		UpdatedAt:     time.Now(), //valueMap["updated_at"].(time.Time),
		UpdatedBy:     StringOrEmpty(valueMap["updated_by"]),
	}
	return m
}
