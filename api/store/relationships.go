package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/pennsieve/model-service-serverless/api/models"
	"log"
)

// CreateRelationShips creates a set of relationships between records that are provided by the user.
// This function accepts a single path (FROM - REL -> TO) and a set of matching pairs of record Ids.
func (s *graphStore) CreateRelationShips(datasetId int, organizationId int, userId string,
	q models.PostRecordRelationshipRequestBody) ([]models.ShortRecordRelationShip, error) {

	ctx := context.Background()

	// Assert that to, and from arrays are the same length
	// This means we can create one relationship between each row from both arrays.
	if len(q.Records.From) != len(q.Records.To) {
		return nil, errors.New("from and To arrays need to be of the same length")
	}

	// 1. CHECK THE MODEL PATH AND RETURN IDS FOR MODELS
	// Match to/from models that are connected by relationship that
	// belong to current dataset and organization and return ids.
	cql := "MATCH (m0:Model{name: $fromModelName})-[r0:`@RELATED_TO`{display_name: $relName}]-" +
		"(m1:Model{name: $toModelName})-[:`@IN_DATASET`]->(:Dataset { id: $datasetID})- " +
		"[:`@IN_ORGANIZATION`]->(:Organization {id: $organizationId}) " +
		"RETURN m1.id AS toID, r0.id AS relID, r0.type AS relType, startnode(r0).id AS startNode, m0.id AS fromID"

	params := map[string]interface{}{
		"toModelName":    q.Relationship.ToModel,
		"relName":        q.Relationship.RelName,
		"fromModelName":  q.Relationship.FromModel,
		"datasetID":      datasetId,
		"organizationId": organizationId,
	}

	results, err := s.db.Run(ctx, cql, params)
	if err != nil {
		log.Printf("Error with running the NEO4J Path query: %s", err)
		return nil, err
	}

	record, err := results.Single(ctx)
	if err != nil {
		userError := errors.New(fmt.Sprintf("Unable to find provided relationship %s-%s-%s in dataset: \n",
			q.Relationship.FromModel, q.Relationship.RelName, q.Relationship.ToModel))
		return nil, userError
	}

	fromID, _ := record.Get("fromID")
	toID, _ := record.Get("toID")
	relId, _ := record.Get("relID")
	relType, _ := record.Get("relType")
	startNode, _ := record.Get("startNode")

	// 2. CHECK THAT PROVIDED RECORDS EXIST IN THE PROVIDED MODELS
	// Match all records that belong to the given models and that are part
	// of the list of records that we want to link. The returned list
	// should contain the same records as provided in the request.
	cql2 := "MATCH (r0:Record)-[:`@INSTANCE_OF`]->(m:Model{id: $FromModelId})" +
		"USING INDEX r0:Record(`@id`) WHERE r0.`@id` IN $RecordFromList " +
		"RETURN DISTINCT r0.`@id` AS id " +
		"UNION ALL " +
		"MATCH (r1:Record)-[:`@INSTANCE_OF`]->(m:Model{id: $ToModelId})" +
		"USING INDEX r1:Record(`@id`) WHERE r1.`@id` IN $RecordToList " +
		"RETURN DISTINCT r1.`@id` AS id"

	params = map[string]interface{}{
		"FromModelId":    fromID,
		"ToModelId":      toID,
		"RecordFromList": q.Records.From,
		"RecordToList":   q.Records.To,
	}

	log.Println(params)

	result, err := s.db.Run(ctx, cql2, params)
	if err != nil {
		log.Println("Error running match records query: ", err)
		return nil, err
	}

	records, err := result.Collect(ctx)
	if err != nil {
		log.Println("Error collecting results from NEO4J query: ", err)
		return nil, err
	}

	// Check if all records in request are valid.
	recordSet := newSet(records, "id")

	for _, r := range append(q.Records.From, q.Records.To...) {
		if !recordSet.has(r) {
			return nil, errors.New(fmt.Sprintf("Not all provided record IDs are present in the dataset: %s", r))
		}
	}

	// 3. CREATE RELATIONSHIPS BETWEEN RECORDS

	// If startNode of relationship is the toID that was provided
	// then switch to, from nodes.
	originNodes := q.Records.From
	targetNodes := q.Records.To
	if startNode == toID {
		originNodes = q.Records.To
		targetNodes = q.Records.From
	}

	cql3 := "UNWIND $batch AS row " +
		"MATCH (from:Record{`@id`: row.from}) " +
		"MATCH (to:Record{`@id`: row.to}) " +
		fmt.Sprintf("MERGE (from)-[rel:%s]->(to) ", relType) +
		"ON CREATE SET rel.created_by = $user, rel.created_at = datetime({timezone:\"Greenwich\"}), " +
		"rel.updated_by = $user, rel.updated_at = datetime({timezone:\"Greenwich\"})," +
		"rel.model_relationship_id = $modelRelID, rel.id = row.uuid " +
		"ON MATCH SET rel.updated_by = $user, rel.updated_at = datetime({timezone:\"Greenwich\"}) " +
		"RETURN rel.id AS relID, row.from AS from, row.to AS to"

	params = map[string]interface{}{
		"batch":      mapNodes(originNodes, targetNodes),
		"user":       userId,
		"modelRelID": relId,
	}

	transaction, err := s.db.BeginTransaction(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txResult, err := transaction.Run(ctx, cql3, params)
	if err != nil {
		return nil, err
	}

	txRecords, err := txResult.Collect(ctx)
	if err != nil {
		return nil, err
	}

	var relationships []models.ShortRecordRelationShip
	for i, _ := range txRecords {
		relationships = append(relationships, parseRecordRelationshipResponse(txRecords[i], relType.(string)))
	}

	err = transaction.Commit(ctx)
	if err != nil {
		return nil, err
	}

	transaction.Close(ctx)
	return relationships, nil

}
