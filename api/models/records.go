package models

import "time"

type PostRecordRelationshipRequestBody struct {
	Relationship ModelRelationShip `json:"relationship"`
	Records      ToFromList        `json:"records"`
}

type ModelRelationShip struct {
	ToModel   string `json:"to_model"`
	RelName   string `json:"relationship_name"`
	FromModel string `json:"from_model"`
}

type ToFromList struct {
	To   []string `json:"to"`
	From []string `json:"from"`
}

type ShortRecordRelationShip struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	To      string `json:"to"`
	RelType string `json:"type"`
}

type RecordRelationShip struct {
	ID         string    `json:"id"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	RelType    string    `json:"type"`
	ModelRelID string    `json:"model_relationship_id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"createdAt"`
	CreatedBy  string    `json:"createdBy"`
	UpdatedAt  time.Time `json:"updatedAt"`
	UpdatedBy  string    `json:"updatedBy"`
}

type Record struct {
	ID    string                 `json:"id"`
	Model string                 `json:"model"`
	Props map[string]interface{} `json:"props"`
}
