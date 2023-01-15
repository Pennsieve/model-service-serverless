package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/pennsieve/model-service-serverless/api/models"
	"log"
	"regexp"
	"strings"
	"time"
)

func (s *graphStore) GetModelByName(modelName string, datasetId int, organizationId int) (*models.Model, error) {

	cql := fmt.Sprintf("MATCH  (m:Model{name:'%s'})", modelName) +
		fmt.Sprintf("-[:`@IN_DATASET`]->(:Dataset { id: %d }) ", datasetId) +
		"MATCH (m)-[created:`@CREATED_BY`]->(c:User)" +
		"MATCH (m)-[updated:`@UPDATED_BY`]->(u:User)" +
		"OPTIONAL MATCH (m)-[r:`@RELATED_TO`]->(n) WHERE r.index IS NOT NULL " +
		"RETURN m.name AS name, m.description AS description, m.id AS id, m.display_name AS display_name," +
		"	size(()-[:`@INSTANCE_OF`]->(m)) AS count," +
		"	size((m)-[:`@HAS_PROPERTY`]->()) AS nrStaticProps, count((m)--(n)) AS nrLinkedProps," +
		"	c.node_id AS created_by, u.node_id AS updated_by, created.at AS created_at, updated.at AS updated_at"

	ctx := context.Background()
	transaction, err := s.db.BeginTransaction(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	result, err := transaction.Run(ctx, cql, nil)
	if err != nil {
		return nil, err
	}

	// Iterate over results and create array of model objects
	record, err := result.Single(ctx)
	if err != nil {
		return nil, err
	}

	m := parseModelResponse(record)
	return &m, nil

}

// InitOrgAndDataset ensures that the metadata database has records for the organization and dataset.
func (s *graphStore) InitOrgAndDataset(organizationId int, datasetId int, organizationNodeId string, datasetNodeId string) error {

	ctx := context.Background()

	cql := "MERGE (o:Organization{id: $organizationId})" +
		"ON CREATE SET o.id = toInteger($organizationId), o.node_id = $organizationNodeId " +
		"ON MATCH SET o.node_id = COALESCE($organizationNodeId, o.node_id) " +
		"MERGE (o)<-[:`@IN_ORGANIZATION`]-(d:Dataset{id: $datasetId}) " +
		"ON CREATE SET d.id = toInteger($datasetId), d.node_id = $datasetNodeId " +
		"ON MATCH SET d.node_id = COALESCE($datasetNodeId, d.node_id) " +
		"RETURN o.id AS organizationId, d.id AS datasetId, o.node_id " +
		"AS organizationNodeId, d.node_id AS datasetNodeId"

	params := map[string]interface{}{
		"organizationId":     organizationId,
		"organizationNodeId": organizationNodeId,
		"datasetId":          datasetId,
		"datasetNodeId":      datasetNodeId,
	}

	results, err := s.db.Run(ctx, cql, params)
	if err != nil {
		log.Printf("Error with running the NEO4J Path query: %s", err)
		return err
	}

	_, err = results.Single(ctx)
	if err != nil {
		return err
	}

	return nil
}

// CreateModel creates a model in a dataset with a specific name and description
func (s *graphStore) CreateModel(datasetId int, organizationId int, name string, displayName string, description string, userId string) (*models.Model, error) {

	// Check if reserved model name
	reservedModelNames := []string{"file"}
	if stringInSlice(name, reservedModelNames) {
		return nil, errors.New(fmt.Sprintf("%s is a reserved name. Unable to create model", name))
	}

	// Validate Model Name
	name, err := validateModelName(name)
	if err != nil {
		return nil, err
	}

	// Check if model with name already exist in dataset
	var cql strings.Builder
	cql.WriteString(fmt.Sprintf("MATCH (m:Model{name: '%s'})", name))
	cql.WriteString(fmt.Sprintf("-[`@IN_DATASET`]->(:Dataset{id:%d})", datasetId))
	cql.WriteString(fmt.Sprintf("-[`@IN_ORGANIZATION`]->(:Organization{id:%d})", organizationId))
	cql.WriteString("RETURN COUNT(m) AS count")

	ctx := context.Background()
	tx, err := s.db.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tx.Run(ctx, cql.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := result.Single(context.Background())
	if err != nil {
		return nil, err
	}

	cnt, exists := res.Get("count")
	if !exists {
		return nil, errors.New("COunt does not exist")
	}

	// Check that model with that name does not exist.
	if cnt.(int64) != 0 {
		return nil, &models.ModelNameCountError{Name: name}
	}

	cql.Reset()

	// MATCHING
	cql.WriteString(fmt.Sprintf("MATCH (d:Dataset{id: %d})-[:`@IN_ORGANIZATION`]->(Organization{id: %d}) ", datasetId, organizationId))
	cql.WriteString(fmt.Sprintf("MERGE (u:User{node_id:'%s'}) ", userId))
	cql.WriteString(fmt.Sprintf("CREATE (m:Model{`@max_sort_key`:0, id:randomUUID(),name:'%s',display_name:'%s',description:'%s'}) ", name, displayName, description))
	cql.WriteString("CREATE (m)-[:`@IN_DATASET`]->(d) ")
	cql.WriteString("CREATE (m)-[created:`@CREATED_BY` {at: datetime()}]->(u) ")
	cql.WriteString("CREATE (m)-[updated:`@UPDATED_BY` {at: datetime()}]->(u) ")
	cql.WriteString("RETURN m, created.at AS created_at, updated.at AS updated_at")

	result, err = tx.Run(ctx, cql.String(), nil)
	if err != nil {
		return nil, err
	}

	// Create key/value map based on keys and values returned.
	valueMap := make(map[string]interface{})
	values := result.Record().Values
	for i, k := range result.Record().Keys {
		valueMap[k] = values[i]
	}

	log.Println(values)

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

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

// GetModels returns a list of models for a provided dataset within an organization
func (s *graphStore) GetModels(datasetId int, organizationId int) (map[string]models.Model, error) {

	var cql strings.Builder

	// MATCHING
	cql.WriteString("MATCH  (m:Model)")
	cql.WriteString(fmt.Sprintf("-[:`@IN_DATASET`]->(:Dataset { id: %d }) ", datasetId))
	cql.WriteString(fmt.Sprintf("-[:`@IN_ORGANIZATION`]->(:Organization { id: %d }) ", organizationId))
	cql.WriteString("MATCH (m)-[created:`@CREATED_BY`]->(c:User)")
	cql.WriteString("MATCH (m)-[updated:`@UPDATED_BY`]->(u:User)")
	cql.WriteString("OPTIONAL MATCH (m)-[r:`@RELATED_TO`]->(n) WHERE r.index IS NOT NULL ")

	// RETURNING
	cql.WriteString("RETURN m.name AS name, m.description AS description, m.id AS id, m.display_name AS display_name,")
	cql.WriteString("size(()-[:`@INSTANCE_OF`]->(m)) AS count,")
	cql.WriteString("size((m)-[:`@HAS_PROPERTY`]->()) AS nrStaticProps, count((m)--(n)) AS nrLinkedProps,")
	cql.WriteString("c.node_id AS created_by, u.node_id AS updated_by, created.at AS created_at, updated.at AS updated_at")

	ctx := context.Background()
	transaction, err := s.db.BeginTransaction(ctx)
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
	modelMap := make(map[string]models.Model)
	for result.Next(ctx) {
		record := parseModelResponse(result.Record())
		modelMap[record.Name] = record
	}
	// Err returns the error that caused Next to return false
	if err = result.Err(); err != nil {
		return nil, err
	}

	transaction.Close(ctx)

	return modelMap, nil

}

// GetModelProps returns a list of properties associated with the provided model
func (s *graphStore) GetModelProps(datasetId int, organizationId int, modelName string) ([]models.ModelProperty, error) {

	var cql strings.Builder

	// MATCHING
	cql.WriteString(fmt.Sprintf("MATCH  (p:ModelProperty)<-[:`@HAS_PROPERTY`]-(:Model { name:'%s' })", modelName))
	cql.WriteString(fmt.Sprintf("-[:`@IN_DATASET`]->(:Dataset { id: %d }) ", datasetId))
	cql.WriteString(fmt.Sprintf("-[:`@IN_ORGANIZATION`]->(:Organization { id: %d }) ", organizationId))

	// RETURN
	cql.WriteString("RETURN p.name AS name, p.description AS description, p.id AS id, p.display_name AS display_name,")
	cql.WriteString(" p.default AS default, p.data_type AS data_type, p.model_title AS model_title, p.index AS index")
	//cql.WriteString(" SORT BY p.index")

	ctx := context.Background()
	transaction, err := s.db.BeginTransaction(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	result, err := transaction.Run(ctx, cql.String(), nil)
	if err != nil {
		return nil, err
	}

	// Iterate over results and create array of model objects
	var modelArr []models.ModelProperty
	for result.Next(ctx) {
		record := parseModelPropertyResponse(result.Record())
		modelArr = append(modelArr, record)
	}
	// Err returns the error that caused Next to return false
	if err = result.Err(); err != nil {
		return nil, err
	}

	transaction.Close(ctx)

	return modelArr, nil

}

// validateModelName returns a valid ModelName or error.
func validateModelName(name string) (string, error) {
	name = strings.TrimSpace(name)
	n := len(name)

	if n == 0 {
		return "", &models.EmptyError{}
	}
	if n > 64 {
		return "", &models.NameTooLongError{Name: name}
	}

	// Will match all unicode letters, but not digits, and underscore "_",
	// followed by letters, digits, and underscores:
	r := regexp.MustCompile(`^([^\W\d_\-]|_)[\w_]*$`)
	isValid := r.MatchString(name)
	if !isValid {
		return "", &models.ValidationError{Name: name}
	}

	return name, nil
}
