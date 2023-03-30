package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/pennsieve/model-service-serverless/api/models"
	log "github.com/sirupsen/logrus"
	"strings"
)

// Query returns an array of records based on a set of filters within a dataset
func (s *graphStore) Query(datasetId int, organizationId int, q models.QueryRequestBody) ([]models.Record, error) {

	modelMap, err := s.GetModels(datasetId, organizationId)
	if err != nil {
		log.Println(err)
	}

	sourceModel, inMap := modelMap[q.Model]
	if inMap == false {
		return nil, &models.UnknownModelError{Model: q.Model}
	}

	// Use default ordering unless specifically defined
	orderBy := q.OrderBy
	if orderBy == "" {
		orderBy = "`@sort_key`"
	} else {
		//	Check if provided value is valid.
		modelProps, err := s.GetModelProps(datasetId, organizationId, q.Model)
		if err != nil {
			return nil, err
		}

		propFound := false
		for _, v := range modelProps {
			if v.Name == q.OrderBy {
				orderBy = q.OrderBy
				propFound = true
				break
			}
		}

		if !propFound {
			return nil, &models.UnknownModelPropertyError{PropName: q.OrderBy}
		}
	}

	ctx := context.Background()

	targetModels, err := getTargetModelsMap(q.Filters, sourceModel, modelMap)

	shortestPaths, err := s.ShortestPath(ctx, sourceModel, targetModels)

	query, err := generateQuery(sourceModel, shortestPaths, q.Filters, orderBy, false, "", "")

	result, err := s.db.Run(ctx, query, nil)
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
			Model: q.Model,
			Props: node.Props,
		}
		records = append(records, newRec)

	}

	return records, nil
}

// Autocomplete returns a list of terms that match values given the specified filters for a property in a model
func (s *graphStore) Autocomplete(datasetId int, organizationId int, q models.AutocompleteRequestBody) ([]string, error) {

	modelMap, err := s.GetModels(datasetId, organizationId)
	if err != nil {
		log.Println(err)
	}

	sourceModel, inMap := modelMap[q.Model]
	if inMap == false {
		return nil, &models.UnknownModelError{Model: q.Model}
	}

	// Use default ordering
	orderBy := "`@sort_key`"

	ctx := context.Background()

	targetModels, err := getTargetModelsMap(q.Filters, sourceModel, modelMap)

	shortestPaths, err := s.ShortestPath(ctx, sourceModel, targetModels)

	query, err := generateQuery(sourceModel, shortestPaths, q.Filters, orderBy, true, q.Text, q.Property)

	log.Println(query)

	result, err := s.db.Run(ctx, query, nil)
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
func (s *graphStore) ShortestPath(ctx context.Context, sourceModel models.Model, targetModels map[string]string) ([]dbtype.Path, error) {

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

	result, err := s.db.Run(ctx, cql, nil)
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
func (s *graphStore) GetRecordsForPackage(ctx context.Context, datasetId int, organizationId int, packageNodeId string, maxDepth int) ([]models.Record, error) {
	// MATCH p = (n:Package{package_node_id:'N:package:ab16ccc2-c5a5-4ead-a476-bc3f6e919364'})<-[*0..5]-(b:Record)-[:`@INSTANCE_OF`]->(m:Model)-[:`@IN_DATASET`]->(:Dataset)-[:`@IN_ORGANIZATION`]->(:Organization) RETURN DISTINCT b as records ,m.name as models

	cql := fmt.Sprintf("MATCH (p:Package{package_node_id:'%s'})-[:`@IN_PACKAGE`]-(a:Record)", packageNodeId) +
		fmt.Sprintf("<-[*0..%d]-(r:Record)--(m:Model)", maxDepth) +
		fmt.Sprintf("-[:`@IN_DATASET`]->(:Dataset { id: %d })-[:`@IN_ORGANIZATION`]->(:Organization { id: %d }) ", datasetId, organizationId) +
		"RETURN DISTINCT r as records ,m.name as models"

	result, err := s.db.Run(ctx, cql, nil)
	if err != nil {
		return nil, err
	}

	var records []models.Record
	for result.Next(ctx) {
		r := result.Record()
		rn, exists := r.Get("records")
		if !exists {
			return nil, errors.New("records not returned from neo4j")
		}

		node := rn.(dbtype.Node)

		mn, exists := r.Get("model")
		if !exists {
			return nil, errors.New("records not returned from neo4j")
		}
		model := mn.(string)

		id := node.Props["@id"].(string)

		// Delete internal properties from map
		delete(node.Props, "@id")
		delete(node.Props, "@sort_key")

		newRec := models.Record{
			ID:    id,
			Model: model,
			Props: node.Props,
		}
		records = append(records, newRec)

	}

	return records, nil
}

// getTargetModelsMap returns a map between model name and model object
func getTargetModelsMap(filters []models.Filters, sourceModel models.Model, modelMap map[string]models.Model) (map[string]string, error) {

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
func generateQuery(sourceModel models.Model, paths []dbtype.Path, filters []models.Filters,
	orderByProp string, autocomplete bool, autocompleteText string, autocompletePropName string) (string, error) {

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
	if autocomplete {
		// Add autocomplete filter
		if !firstWhereClause {
			queryStr.WriteString("AND ")
		} else {
			queryStr.WriteString("WHERE ")
		}

		queryStr.WriteString(fmt.Sprintf("%s.%s =~ '(?i).*%s.*' ", sourceModel.Name, autocompletePropName, autocompleteText))
		queryStr.WriteString(fmt.Sprintf("RETURN DISTINCT %s.%s AS value LIMIT %d", sourceModel.Name, autocompletePropName, 20))
	} else {
		queryStr.WriteString(fmt.Sprintf("RETURN DISTINCT %s AS records ORDER BY %s.%s LIMIT %d", sourceModel.Name, sourceModel.Name, orderByProp, 100))
	}

	return queryStr.String(), nil
}

// validOperator checks if the requested operator is one of the allowed methods.
func validOperator(op string) bool {
	var validOperators = [...]string{
		"IS", "IS NOT", "EQUALS", "NOT EQUALS", "LESS THAN", "LESS THAN EQUALS",
		"GREATER THAN", "GREATER THAN EQUALS", "STARTS WITH", "CONTAINS"}

	for _, o := range validOperators {
		if op == o {
			return true
		}
	}
	return false
}
