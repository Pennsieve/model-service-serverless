package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/pennsieve/model-service-serverless/api/models"
	"github.com/pennsieve/model-service-serverless/api/models/query"
	"github.com/pennsieve/model-service-serverless/api/shared"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DB interface for queries which has methods that are both available in db connection and in transaction.
type DB interface {
	Run(ctx context.Context, cypher string, params map[string]any) (neo4j.ResultWithContext, error)
}

type NeoQueries struct {
	db DB
}

func NewNeoQueries(db DB) *NeoQueries {
	return &NeoQueries{
		db: db,
	}
}

func (q *NeoQueries) GetModelByName(ctx context.Context, modelName string, datasetId int) (*models.Model, error) {

	cql := fmt.Sprintf("MATCH  (m:Model{name:'%s'})", modelName) +
		fmt.Sprintf("-[:`@IN_DATASET`]->(:Dataset { id: %d }) ", datasetId) +
		"MATCH (m)-[created:`@CREATED_BY`]->(c:User)" +
		"MATCH (m)-[updated:`@UPDATED_BY`]->(u:User)" +
		"OPTIONAL MATCH (m)-[r:`@RELATED_TO`]->(n) WHERE r.index IS NOT NULL " +
		"RETURN m.name AS name, m.description AS description, m.id AS id, m.display_name AS display_name," +
		"	size(()-[:`@INSTANCE_OF`]->(m)) AS count," +
		"	size((m)-[:`@HAS_PROPERTY`]->()) AS nrStaticProps, count((m)--(n)) AS nrLinkedProps," +
		"	c.node_id AS created_by, u.node_id AS updated_by, created.at AS created_at, updated.at AS updated_at"

	result, err := q.db.Run(ctx, cql, nil)
	if err != nil {
		return nil, err
	}

	// Iterate over results and create array of model objects
	record, err := result.Single(ctx)
	if err != nil {
		return nil, err
	}

	m := shared.ParseModelResponse(record)
	return &m, nil

}

// InitOrgAndDataset ensures that the metadata database has records for the organization and dataset.
func (q *NeoQueries) InitOrgAndDataset(ctx context.Context, organizationId int, datasetId int, organizationNodeId string, datasetNodeId string) error {

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

	results, err := q.db.Run(ctx, cql, params)
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
func (q *NeoQueries) CreateModel(ctx context.Context, datasetId int, organizationId int, name string, displayName string, description string, userId string) (*models.Model, error) {

	// Check if reserved model name
	reservedModelNames := []string{"file"}
	if shared.StringInSlice(name, reservedModelNames) {
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

	result, err := q.db.Run(ctx, cql.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := result.Single(context.Background())
	if err != nil {
		return nil, err
	}

	cnt, exists := res.Get("count")
	if !exists {
		return nil, errors.New("Count does not exist")
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

	result, err = q.db.Run(ctx, cql.String(), nil)
	if err != nil {
		return nil, err
	}

	rec, err := result.Single(ctx)
	if err != nil {
		return nil, err
	}

	m, _ := rec.Get("m")
	modelNode := m.(neo4j.Node)

	c, _ := rec.Get("created_at")
	createdAt := c.(time.Time)

	u, _ := rec.Get("created_at")
	updatedAt := u.(time.Time)

	mo := models.Model{
		Count:       0,
		CreatedAt:   createdAt,
		CreatedBy:   userId,
		Description: shared.StringOrEmpty(modelNode.Props["description"]),
		DisplayName: shared.StringOrEmpty(modelNode.Props["display_name"]),
		ID:          shared.StringOrEmpty(modelNode.Props["id"]),
		Locked:      false,
		Name:        shared.StringOrEmpty(modelNode.Props["name"]),
		TemplateID:  nil,
		UpdatedAt:   updatedAt,
		UpdatedBy:   userId,
	}

	return &mo, nil
}

// GetModels returns a list of models for a provided dataset within an organization
func (q *NeoQueries) GetModels(ctx context.Context, datasetId int, organizationId int) (map[string]models.Model, error) {

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

	result, err := q.db.Run(ctx, cql.String(), nil)
	if err != nil {
		return nil, err
	}

	// Iterate over results and create array of model objects
	modelMap := make(map[string]models.Model)
	for result.Next(ctx) {
		record := shared.ParseModelResponse(result.Record())
		modelMap[record.Name] = record
	}
	// Err returns the error that caused Next to return false
	if err = result.Err(); err != nil {
		return nil, err
	}

	return modelMap, nil

}

// GetModelProps returns a list of properties associated with the provided model
func (q *NeoQueries) GetModelProps(ctx context.Context, datasetId int, organizationId int, modelName string) ([]models.ModelProperty, error) {

	var cql strings.Builder

	// MATCHING
	cql.WriteString(fmt.Sprintf("MATCH  (p:ModelProperty)<-[:`@HAS_PROPERTY`]-(:Model { name:'%s' })", modelName))
	cql.WriteString(fmt.Sprintf("-[:`@IN_DATASET`]->(:Dataset { id: %d }) ", datasetId))
	cql.WriteString(fmt.Sprintf("-[:`@IN_ORGANIZATION`]->(:Organization { id: %d }) ", organizationId))

	// RETURN
	cql.WriteString("RETURN p.name AS name, p.description AS description, p.id AS id, p.display_name AS display_name,")
	cql.WriteString(" p.default AS default, p.data_type AS data_type, p.model_title AS model_title, p.index AS index")
	//cql.WriteString(" SORT BY p.index")

	result, err := q.db.Run(ctx, cql.String(), nil)
	if err != nil {
		return nil, err
	}

	// Iterate over results and create array of model objects
	var modelArr []models.ModelProperty
	for result.Next(ctx) {
		record := shared.ParseModelPropertyResponse(result.Record())
		modelArr = append(modelArr, record)
	}
	// Err returns the error that caused Next to return false
	if err = result.Err(); err != nil {
		return nil, err
	}

	return modelArr, nil

}

// QueryTotal returns the total number of results for a particular query
func (q *NeoQueries) QueryTotal(ctx context.Context, sourceModel models.Model, shortestPaths []dbtype.Path, filters []query.Filters,
	orderBy string, limit int, offset int) (int, error) {

	queryParams := query.FormatParams{ResultType: query.COUNT}

	query, err := generateQuery(sourceModel, shortestPaths, filters, orderBy, queryParams, limit, offset)
	if err != nil {
		log.Error("Error generating query: ", err)
		return 0, err
	}

	log.Debug("Query: ", query)

	result, err := q.db.Run(ctx, query, nil)
	if err != nil {
		return 0, err
	}

	record, err := result.Single(ctx)
	if err != nil {
		return 0, err
	}
	r, exists := record.Get("total")
	if !exists {
		return 0, errors.New("result does not contain property 'total'")
	}

	totalNrRecords := r.(int)
	return totalNrRecords, nil
}

// Query returns an array of records based on a set of filters within a dataset
func (q *NeoQueries) Query(ctx context.Context, sourceModel models.Model, shortestPaths []dbtype.Path, filters []query.Filters,
	orderBy string, limit int, offset int) ([]models.Record, error) {

	queryParams := query.FormatParams{ResultType: query.RESULTS}
	query, err := generateQuery(sourceModel, shortestPaths, filters, orderBy, queryParams, limit, offset)
	if err != nil {
		log.Error("Error generating query: ", err)
		return nil, err
	}

	log.Debug("Query: ", query)

	result, err := q.db.Run(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	var records []models.Record
	for result.Next(ctx) {
		r := result.Record()
		rn, _ := r.Get("records")
		node := rn.(dbtype.Node)

		id := node.Props["@id"].(string)

		// Delete internal properties from map
		delete(node.Props, "@id")
		delete(node.Props, "@sort_key")

		newRec := models.Record{
			ID:    id,
			Model: sourceModel.Name,
			Props: node.Props,
		}
		records = append(records, newRec)

	}

	return records, nil
}

// Autocomplete returns a list of terms that match values given the specified filters for a property in a model
func (q *NeoQueries) Autocomplete(ctx context.Context, datasetId int, organizationId int, req query.AutocompleteRequestBody) ([]string, error) {

	modelMap, err := q.GetModels(ctx, datasetId, organizationId)
	if err != nil {
		log.Println(err)
	}

	sourceModel, inMap := modelMap[req.Model]
	if inMap == false {
		return nil, &models.UnknownModelError{Model: req.Model}
	}

	// Use default ordering
	orderBy := "`@sort_key`"

	targetModels, err := getTargetModelsMap(req.Filters, sourceModel, modelMap)

	shortestPaths, err := q.ShortestPath(ctx, sourceModel, targetModels)

	params := query.FormatParams{
		ResultType: query.AUTOCOMPLETE,
		AutoCompleteParams: query.AutoCompleteParams{
			Text:     req.Text,
			PropName: req.Property,
		},
	}

	query, err := generateQuery(sourceModel, shortestPaths, req.Filters, orderBy, params, 20, 0)

	log.Debug(query)

	result, err := q.db.Run(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	records, err := result.Collect(ctx)
	var values []string
	if err != nil {
		return values, err
	}

	for _, v := range records {
		value, _ := v.Get("value")
		values = append(values, value.(string))
	}

	return values, nil
}

// ShortestPath returns the shortest paths between source model and the target models
func (q *NeoQueries) ShortestPath(ctx context.Context, sourceModel models.Model, targetModels map[string]string) ([]dbtype.Path, error) {

	keys := make([]string, len(targetModels))
	i := 0
	for _, v := range targetModels {
		keys[i] = v
		i++
	}

	targetModelStr := fmt.Sprintf("['%s']", strings.Join(keys, "','"))

	cql := fmt.Sprintf("MATCH (m:Model{id:'%s'})-[:`@IN_DATASET`]->(d:Dataset)-[:`@IN_ORGANIZATION`]->(o:Organization) ", sourceModel.ID) +
		"MATCH (n:Model)-[:`@IN_DATASET`]->(d) " +
		fmt.Sprintf("WHERE n.id IN %s ", targetModelStr) +
		"MATCH p = shortestPath((m)-[:`@RELATED_TO` *..4]-(n)) " +
		"RETURN p AS path"

	result, err := q.db.Run(ctx, cql, nil)
	if err != nil {
		return nil, err
	}

	var shortestPaths []dbtype.Path
	for result.Next(ctx) {
		r := result.Record()
		path, _ := r.Get("path")
		shortestPaths = append(shortestPaths, path.(dbtype.Path))

	}

	return shortestPaths, nil
}

// GetRecordsForPackage returns a list of connected records
func (q *NeoQueries) GetRecordsForPackage(ctx context.Context, datasetId int, organizationId int, packageIds []int, maxDepth int) ([]models.PackageMetadata, error) {
	// MATCH p = (n:Package{package_node_id:'N:package:ab16ccc2-c5a5-4ead-a476-bc3f6e919364'})<-[*0..5]-(b:Record)-[:`@INSTANCE_OF`]->(m:Model)-[:`@IN_DATASET`]->(:Dataset)-[:`@IN_ORGANIZATION`]->(:Organization) RETURN DISTINCT b as records ,m.name as models

	//cql := fmt.Sprintf("MATCH (p:Package{package_node_id:'%s'})-[:`@IN_PACKAGE`]-(a:Record)", packageNodeId) +
	//	fmt.Sprintf("<-[*0..%d]-(r:Record)--(m:Model)", maxDepth) +
	//	fmt.Sprintf("-[:`@IN_DATASET`]->(:Dataset { id: %d })-[:`@IN_ORGANIZATION`]->(:Organization { id: %d }) ", datasetId, organizationId) +
	//	"RETURN DISTINCT r as records ,m.name as model"

	var IDs []string
	for _, i := range packageIds {
		IDs = append(IDs, strconv.Itoa(i))
	}
	ancestorIds := strings.Join(IDs, ",")

	cql := "" +
		fmt.Sprintf("MATCH (p:Package)<-[*0..%d]-(r:Record)-", maxDepth) +
		fmt.Sprintf("[:`@INSTANCE_OF`]->(m:Model)-[:`@IN_DATASET`]->(:Dataset{id: %d })-[:`@IN_ORGANIZATION`]->(:Organization{id: %d }) ", datasetId, organizationId) +
		fmt.Sprintf("WHERE p.package_id IN [%s] RETURN DISTINCT r as records ,m.name as model, {node_id:p.package_node_id, id:p.package_id} AS origin", ancestorIds)

	result, err := q.db.Run(ctx, cql, nil)

	if err != nil {
		return nil, err
	}

	var records []models.PackageMetadata
	for result.Next(ctx) {
		r := result.Record()
		rn, exists := r.Get("records")
		if !exists {
			return nil, errors.New("records not returned from neo4j")
		}

		node := rn.(dbtype.Node)

		mn, exists := r.Get("model")
		if !exists {
			return nil, errors.New("model not returned from neo4j")
		}
		model := mn.(string)

		mo, exists := r.Get("origin")
		if !exists {
			return nil, errors.New("origin not returned from neo4j")
		}
		or := mo.(map[string]interface{})
		origin := models.OriginRecord{
			Id:     or["id"].(int64),
			NodeId: or["node_id"].(string),
		}

		id := node.Props["@id"].(string)

		// Delete internal properties from map
		delete(node.Props, "@id")
		delete(node.Props, "@sort_key")

		newRec := models.PackageMetadata{
			ID:     id,
			Model:  model,
			Props:  node.Props,
			Origin: origin,
		}
		records = append(records, newRec)

	}

	return records, nil
}

// CreateRelationShips creates a set of relationships between records that are provided by the user.
// This function accepts a single path (FROM - REL -> TO) and a set of matching pairs of record Ids.
func (q *NeoQueries) CreateRelationShips(ctx context.Context, datasetId int, organizationId int, userId string,
	req models.PostRecordRelationshipRequestBody) ([]models.ShortRecordRelationShip, error) {

	// Assert that to, and from arrays are the same length
	// This means we can create one relationship between each row from both arrays.
	if len(req.Records.From) != len(req.Records.To) {
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
		"toModelName":    req.Relationship.ToModel,
		"relName":        req.Relationship.RelName,
		"fromModelName":  req.Relationship.FromModel,
		"datasetID":      datasetId,
		"organizationId": organizationId,
	}

	results, err := q.db.Run(ctx, cql, params)
	if err != nil {
		log.Printf("Error with running the NEO4J Path query: %s", err)
		return nil, err
	}

	record, err := results.Single(ctx)
	if err != nil {
		userError := errors.New(fmt.Sprintf("Unable to find provided relationship %s-%s-%s in dataset: \n",
			req.Relationship.FromModel, req.Relationship.RelName, req.Relationship.ToModel))
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
		"RecordFromList": req.Records.From,
		"RecordToList":   req.Records.To,
	}

	log.Println(params)

	result, err := q.db.Run(ctx, cql2, params)
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
	recordSet := shared.NewSet(records, "id")

	for _, r := range append(req.Records.From, req.Records.To...) {
		if !recordSet.Has(r) {
			return nil, errors.New(fmt.Sprintf("Not all provided record IDs are present in the dataset: %s", r))
		}
	}

	// 3. CREATE RELATIONSHIPS BETWEEN RECORDS

	// If startNode of relationship is the toID that was provided
	// then switch to, from nodes.
	originNodes := req.Records.From
	targetNodes := req.Records.To
	if startNode == toID {
		originNodes = req.Records.To
		targetNodes = req.Records.From
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
		"batch":      shared.MapNodes(originNodes, targetNodes),
		"user":       userId,
		"modelRelID": relId,
	}

	txResult, err := q.db.Run(ctx, cql3, params)
	if err != nil {
		return nil, err
	}

	txRecords, err := txResult.Collect(ctx)
	if err != nil {
		return nil, err
	}

	var relationships []models.ShortRecordRelationShip
	for i, _ := range txRecords {
		relationships = append(relationships, shared.ParseRecordRelationshipResponse(txRecords[i], relType.(string)))
	}

	return relationships, nil

}

// getTargetModelsMap returns a map between model name and model object
func getTargetModelsMap(filters []query.Filters, sourceModel models.Model, modelMap map[string]models.Model) (map[string]string, error) {

	targetModels := make(map[string]string)
	for _, v := range filters {
		if v.Model != sourceModel.Name {
			tModel, inMap := modelMap[v.Model]
			if !inMap {
				return nil, &models.UnknownModelError{Model: v.Model}
			}
			targetModels[v.Model] = tModel.ID
		}

		// Check operator
		if !validOperator(v.Operator) {
			return nil, &models.UnsupportedOperatorError{Operator: v.Operator}
		}

	}

	return targetModels, nil

}

// generateQuery returns a Cypher query based on the provided paths and filters.
func generateQuery(sourceModel models.Model, paths []dbtype.Path, filters []query.Filters,
	orderByProp string, formatParams query.FormatParams, limit int, offset int) (string, error) {

	if orderByProp == "" {
		return "", errors.New("orderBy cannot be empty")
	}

	// Dynamically build the queryStr
	queryStr := strings.Builder{}
	queryStr.WriteString("MATCH ")

	// Iterate over all shortest paths
	setRestart := false
	if len(paths) > 0 {
		for iPath, p := range paths {

			// Iterate over all nodes within a single path
			for pathIndex := range p.Nodes {
				curNode := p.Nodes[pathIndex]

				// Skip the model node for anything except first path.
				if iPath == 0 {

					if pathIndex == 0 {
						curRel := p.Relationships[pathIndex]
						queryStr.WriteString(fmt.Sprintf("(M%s:Model{id:'%s'})<-[:`@INSTANCE_OF`]-(%s:Record)-[:%s]-",
							curNode.Props["name"], curNode.Props["id"], curNode.Props["name"], curRel.Props["type"]))
					} else {
						if pathIndex <= len(p.Relationships)-1 {
							curRel := p.Relationships[pathIndex]
							queryStr.WriteString(fmt.Sprintf("(%s:Record)-[:%s]-", curNode.Props["name"], curRel.Props["type"]))
						} else {
							queryStr.WriteString(fmt.Sprintf("(%s:Record)-[:`@INSTANCE_OF`]->(M%s:Model{id:'%s'}) ",
								p.Nodes[len(p.Nodes)-1].Props["name"], p.Nodes[len(p.Nodes)-1].Props["name"], p.Nodes[len(p.Nodes)-1].Props["id"]))
							setRestart = true
						}
					}

				} else {

					// Iterate over previous paths to check if the current node is already includeded in the graph-query.
					// We can do this because the paths all start from the same model, and the shortest route to any node
					// does not change between paths.
					pathElExists := false
					for ip, previousPath := range paths {

						// Don't check existing or unprocessed paths
						if ip == iPath {
							break
						}

						// If a previous path has already included the node, mark the node as existing.
						if pathIndex <= len(previousPath.Nodes)-1 {
							if curNode.Props["name"] == previousPath.Nodes[pathIndex].Props["name"] {
								pathElExists = true
							}
						}
					}

					// If the node does not exist in the query, include the element. AND
					// if this is the first element of a new path, then start at the last known element.
					if !pathElExists {
						// In case this is the first node in the path, include the last previously known node as the starting point.
						if setRestart {
							queryStr.WriteString(fmt.Sprintf(", (%s:Record)-[:%s]-", p.Nodes[len(p.Nodes)-2].Props["name"], p.Relationships[len(p.Nodes)-2].Props["type"]))
						}

						// Check if this is the final node, or if this is a waypoint.
						if pathIndex == len(p.Relationships) {
							queryStr.WriteString(fmt.Sprintf("(%s:Record) ", p.Nodes[len(p.Nodes)-1].Props["name"]))
						} else {
							curRel := p.Relationships[pathIndex]
							queryStr.WriteString(fmt.Sprintf("(%s:Record)-[:%s]-", curNode.Props["name"], curRel.Props["type"]))
						}

					}

				}

			}

		}
	} else {
		// No paths; only filters on model that is requested
		queryStr.WriteString(fmt.Sprintf("(M%s:Model{id:'%s'})<-[:`@INSTANCE_OF`]-(%s:Record) ",
			sourceModel.Name, sourceModel.ID, sourceModel.Name))
	}
	// Include WHERE clauses
	firstWhereClause := true
	for _, f := range filters {
		if !firstWhereClause {
			queryStr.WriteString("AND ")
		} else {
			queryStr.WriteString("WHERE ")
		}
		queryStr.WriteString(fmt.Sprintf("%s.%s %s '%s' ", f.Model, f.Property, f.Operator, f.Value))
		firstWhereClause = false
	}

	// Return
	switch formatParams.ResultType {
	case query.AUTOCOMPLETE:
		// Add autocomplete filter
		if !firstWhereClause {
			queryStr.WriteString("AND ")
		} else {
			queryStr.WriteString("WHERE ")
		}

		queryStr.WriteString(fmt.Sprintf("%s.%s =~ '(?i).*%s.*' ", sourceModel.Name,
			formatParams.AutoCompleteParams.PropName, formatParams.AutoCompleteParams.Text))
		queryStr.WriteString(fmt.Sprintf("RETURN DISTINCT %s.%s AS value LIMIT %d",
			sourceModel.Name, formatParams.AutoCompleteParams.PropName, limit))
	case query.RESULTS:
		queryStr.WriteString(fmt.Sprintf("RETURN DISTINCT %s AS records ORDER BY %s.%s SKIP %d LIMIT %d",
			sourceModel.Name, sourceModel.Name, orderByProp, offset, limit))
	case query.COUNT:
		queryStr.WriteString(fmt.Sprintf("RETURN count(distinct %s) AS total", sourceModel.Name))

	}

	return queryStr.String(), nil
}

// validOperator checks if the requested operator is one of the allowed methods.
func validOperator(op string) bool {
	var validOperators = [...]string{
		"=", "<>", "<", "<=",
		">", ">=", "=~", "STARTS WITH", "ENDS WITH", "CONTAINS"}

	for _, o := range validOperators {
		if op == o {
			return true
		}
	}
	return false
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
