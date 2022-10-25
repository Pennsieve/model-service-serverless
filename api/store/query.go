package store

import (
	"context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/pennsieve/model-service-serverless/api/models"
	"log"
	"strings"
)

type SearchPath struct {
	Source string
	Target string
	Path   []PathElement
}

type PathElement struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type,omitempty"`
}

var validOperators = [...]string{"IS", "IS NOT", "EQUALS", "NOT EQUALS", "LESS THAN", "LESS THAN EQUALS", "GREATER THAN", "GREATER THAN EQUALS", "STARTS WITH", "CONTAINS"}

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

	targetModels := make(map[string]string)
	for _, v := range q.Filters {
		if v.Model != q.Model {
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

	ctx := context.Background()

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

	query, err := generateQuery(shortestPaths, q.Filters)

	result, err = s.db.Run(ctx, query, nil)
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

func generateQuery(paths []dbtype.Path, filters []models.Filters) (string, error) {

	// First Path
	sql := strings.Builder{}
	sql.WriteString("MATCH ")

	// Iterate over all shortest paths
	setRestart := false
	for iPath, p := range paths {

		// Iterate over all nodes within a single path

		for pathIndex := range p.Nodes {
			curNode := p.Nodes[pathIndex]

			// Skip the model node for anything except first path.
			if iPath == 0 {

				if pathIndex == 0 {
					curRel := p.Relationships[pathIndex]
					sql.WriteString(fmt.Sprintf("(M%s:Model{id:'%s'})<-[:`@INSTANCE_OF`]-(%s:Record)-[:%s]-",
						curNode.Props["name"], curNode.Props["id"], curNode.Props["name"], curRel.Props["type"]))
				} else {
					if pathIndex <= len(p.Relationships)-1 {
						curRel := p.Relationships[pathIndex]
						sql.WriteString(fmt.Sprintf("(%s:Record)-[:%s]-", curNode.Props["name"], curRel.Props["type"]))
					} else {
						sql.WriteString(fmt.Sprintf("(%s:Record)-[:`@INSTANCE_OF`]->(M%s:Model{id:'%s'}) ",
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
						sql.WriteString(fmt.Sprintf(", (%s:Record)-[:%s]-", p.Nodes[len(p.Nodes)-2].Props["name"], p.Relationships[len(p.Nodes)-2].Props["type"]))
					}

					if pathIndex == len(p.Relationships) {

						sql.WriteString(fmt.Sprintf("(%s:Record) ", p.Nodes[len(p.Nodes)-1].Props["name"]))

					} else {

						curRel := p.Relationships[pathIndex]
						sql.WriteString(fmt.Sprintf("(%s:Record)-[:%s]-", curNode.Props["name"], curRel.Props["type"]))
					}

				}

			}

		}

	}

	// Include WHERE clauses
	firstWhereClause := true
	for _, f := range filters {
		if !firstWhereClause {
			sql.WriteString("AND ")
		} else {
			sql.WriteString("WHERE ")
		}
		sql.WriteString(fmt.Sprintf("%s.%s %s '%s' ", f.Model, f.Property, f.Operator, f.Value))
		firstWhereClause = false
	}

	// Return
	sql.WriteString(fmt.Sprintf("RETURN %s AS records LIMIT %d", paths[0].Nodes[0].Props["name"], 100))

	return sql.String(), nil
}

func validOperator(op string) bool {
	for _, o := range validOperators {
		if op == o {
			return true
		}
	}
	return false
}
