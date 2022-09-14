package models

import (
	"fmt"
	"github.com/pennsieve/model-service-serverless/api/core"
	"log"
	"strings"
)

func GetModels(session core.Neo4jAPI, datasetId int, organizationId int) {

	var cql strings.Builder
	cql.WriteString("MATCH  (m:Model)")
	cql.WriteString(fmt.Sprintf("-[`@IN_DATASET`]->(Dataset { id: %d }) ", datasetId))
	cql.WriteString(fmt.Sprintf("-[`@IN_ORGANIZATION`]->(Organization { id: %d }) ", organizationId))
	cql.WriteString("MATCH (m)-[created:`@CREATED_BY`]->(c:User)")
	cql.WriteString("MATCH (m)-[updated:`@UPDATED_BY`]->(u:User)")
	cql.WriteString("RETURN m, size(()-[`@INSTANCE_OF`]->(m)) AS count, c.node_id AS created_by, ")
	cql.WriteString("u.node_id AS updated_by, created_at AS created_at, updated.at AS updated_at")

	log.Println(cql.String())

}
