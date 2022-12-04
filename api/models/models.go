package models

import (
	"fmt"
	"time"
)

type Model struct {
	Count         int64       `json:"count"`
	CreatedAt     time.Time   `json:"createdAt"`
	CreatedBy     string      `json:"createdBy"`
	Description   string      `json:"description"`
	DisplayName   string      `json:"displayName"`
	ID            string      `json:"id"`
	Locked        bool        `json:"locked"`
	Name          string      `json:"name"`
	PropertyCount int64       `json:"propertyCount"`
	TemplateID    interface{} `json:"templateId"`
	UpdatedAt     time.Time   `json:"updatedAt"`
	UpdatedBy     string      `json:"updatedBy"`
}

type ModelProperty struct {
	ID           string      `json:"id"`
	DataType     interface{} `json:"data_type"`
	DefaultValue interface{} `json:"default"`
	DisplayName  string      `json:"display_name"`
	Name         string      `json:"name"`
	IsModelTitle bool        `json:"model_title"`
	Index        int64       `json:"index"`
}

// String returns a string representation of the model.
func (m Model) String() string {
	return fmt.Sprintf("Model -- name: %s, id: %s", m.Name, m.ID)
}
