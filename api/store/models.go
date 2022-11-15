package store

import (
	"context"
	"fmt"
	"github.com/pennsieve/model-service-serverless/api/models"
	"log"
	"strings"
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
