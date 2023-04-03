package shared

import (
	"context"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jSession returns object that implement neo4j.SessionWithContext but replaces the run method
// to ignore the last optional parameters. Now, we can interchangeably use a Neo4J transaction and direct DB connection.
type Neo4jSession struct {
	s neo4j.SessionWithContext
}

func NewNeo4jSession(session neo4j.SessionWithContext) *Neo4jSession {
	return &Neo4jSession{session}
}

func (n *Neo4jSession) Run(ctx context.Context, cypher string, params map[string]any) (neo4j.ResultWithContext, error) {
	return n.s.Run(ctx, cypher, params)
}

func (n *Neo4jSession) BeginTransaction(ctx context.Context, configurers ...func(*neo4j.TransactionConfig)) (neo4j.ExplicitTransaction, error) {
	return n.s.BeginTransaction(ctx, configurers...)
}

func (n *Neo4jSession) Close(ctx context.Context) error {
	return n.s.Close(ctx)
}
