package store

import (
	"github.com/pennsieve/model-service-serverless/api/models"
	"log"
)

// ** STEP 1 -- Get shortest path to all target nodes
//MATCH (m:Model{id:"5898ec3c-98f4-45c9-a86c-db6f03c6350b"})-[:`@IN_DATASET`]->(d:Dataset)-[:`@IN_ORGANIZATION`]->(o:Organization)
//
//MATCH (n:Model)-[:`@IN_DATASET`]->(d)
//WHERE n.id IN ["5880cac7-d441-4304-bd20-0da9c425ca2f","e476c74d-b3d3-41fe-a558-fd3d3237583e","c5d66575-3f71-42de-81e2-0c8a0a84beba"]
//
//MATCH p = shortestPath((m)-[:`@RELATED_TO` *..4]-(n))
//RETURN p

// ** For each of the models --> match the available records.
// MATCH (m1:Experiment)->

func (s *graphStore) Query(datasetId int, organizationId int, q models.QueryRequestBody) error {

	//var predicates []Predicate

	source, err := s.GetModelByName(q.Model, datasetId, organizationId)
	if err != nil {
		return err
	}
	log.Println(source.String())
	log.Println(source)

	return nil
}
