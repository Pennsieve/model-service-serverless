package store

import (
	"context"
	"github.com/pennsieve/model-service-serverless/api/models"
	pgQueries "github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

// ModelServicePgQueries is the UploadHandler Queries Struct embedding the shared Queries struct
type ModelServicePgQueries struct {
	*pgQueries.Queries
	db pgQueries.DBTX
}

// NewModelServicePgQueries returns a new instance of an ModelServicePgQueries object
func NewModelServicePgQueries(db pgQueries.DBTX) *ModelServicePgQueries {
	q := pgQueries.New(db)
	return &ModelServicePgQueries{
		q,
		db,
	}
}

func (q *ModelServicePgQueries) GetPackageAncestors(ctx context.Context, packageId string) ([]models.PackageAncestor, error) {
	//WITH RECURSIVE folders AS (
	//    SELECT
	//        id,
	//        parent_id,
	//        name,
	//        node_id
	//    FROM packages
	//    WHERE node_id = 'N:package:01862e26-565a-4b1e-a7af-01c3960b0046'
	//    UNION
	//    SELECT
	//        e.id,
	//        e.parent_id,
	//        e.name,
	//        e.node_id
	//    FROM packages e
	//            INNER JOIN folders s ON s.parent_id = e.id
	//) SELECT
	//      *
	//FROM
	//    folders

	var result []models.PackageAncestor

	queryStr := "" +
		"WITH RECURSIVE ancestors AS (" +
		"SELECT " +
		"id, parent_id, name, node_id " +
		"FROM packages " +
		"WHERE node_id = $1 " +
		"UNION SELECT " +
		"e.id, e.parent_id, e.name, e.node_id " +
		"FROM packages e " +
		"INNER JOIN ancestors s ON s.parent_id = e.id) " +
		"SELECT * FROM ancestors"

	log.Println(queryStr)

	rows, err := q.db.QueryContext(ctx, queryStr, packageId)
	if err != nil {
		log.Error("Unable to get ancestors", err)
		return nil, err
	}

	if err == nil {
		for rows.Next() {
			var currentRecord models.PackageAncestor
			err = rows.Scan(
				&currentRecord.Id,
				&currentRecord.ParentId,
				&currentRecord.Name,
				&currentRecord.NodeId)

			if err != nil {
				log.Error("unable to parse ancestor object: ", err)
				return nil, err
			}

			result = append(result, currentRecord)
		}
		return result, err
	}

	// return empty if no ancestors (should not happen)
	return result, nil
}
