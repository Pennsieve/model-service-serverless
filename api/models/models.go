package models

import (
	"context"
	"fmt"
	"github.com/pennsieve/model-service-serverless/service/api/core"
	"log"
	"strings"
)

func getModels(session core.Neo4jAPI, datasetId int, organizationId int) {

	var cql strings.Builder
	cql.WriteString("MATCH  (m:Model)")
	cql.WriteString(fmt.Sprintf("-[`@IN_DATASET`]->(Dataset { id: %d }) ", datasetId))
	cql.WriteString(fmt.Sprintf("-[`@IN_ORGANIZATION`]->(Organization { id: %d }) ", organizationId))
	cql.WriteString("MATCH (m)-[created:`@CREATED_BY`]->(c:User)")
	cql.WriteString("MATCH (m)-[updated:`@UPDATED_BY`]->(u:User)")
	cql.WriteString("RETURN m, size(()-[`@INSTANCE_OF`]->(m)) AS count, c.node_id AS created_by, ")
	cql.WriteString("u.node_id AS updated_by, created_at AS created_at, updated.at AS updated_at")

	log.Println(cql.String())

	transaction, err := session.BeginTransaction(context.Background())
	if err != nil {
		log.Println(err)
		return nil, err
	}

	result, err := transaction.Run(context.Background(),
		"CREATE (a:Greeting) SET a.message = $message RETURN a.message + ', from node ' + id(a)",
		map[string]any{"message": "hello, world"})

	if err != nil {
		return nil, err
	}

	fmt.Println(result)

}
