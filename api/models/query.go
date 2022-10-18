package models

type QueryRequestBody struct {
	Model   string    `json:"model"`
	Filters []Filters `json:"filters"`
}
type Filters struct {
	Model    string `json:"model"`
	Property string `json:"property"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}
