package models

import "database/sql"

type PackageAncestorsResponse struct {
	id        string            `json:"id"`
	ancestors []PackageAncestor `json:"ancestors"`
}

type PackageAncestor struct {
	Id       int64          `json:"id"`
	NodeId   string         `json:"node_id"`
	Name     string         `json:"name"`
	ParentId sql.NullString `json:"parent_id"`
}
