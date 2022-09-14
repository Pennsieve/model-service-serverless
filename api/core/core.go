package core

import (
	"context"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jAPI interface {
	LastBookmarks() neo4j.Bookmarks
	lastBookmark() string
	BeginTransaction(ctx context.Context, configurers ...func(*neo4j.TransactionConfig)) (neo4j.ExplicitTransaction, error)
	ExecuteRead(ctx context.Context, work neo4j.ManagedTransactionWork, configurers ...func(*neo4j.TransactionConfig)) (any, error)
	ExecuteWrite(ctx context.Context, work neo4j.ManagedTransactionWork, configurers ...func(*neo4j.TransactionConfig)) (any, error)
	Run(ctx context.Context, cypher string, params map[string]any, configurers ...func(*neo4j.TransactionConfig)) (neo4j.ResultWithContext, error)
	Close(ctx context.Context) error
}
