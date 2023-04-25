package models

type QueryRequestBody struct {
	Model   string    `json:"model"`
	Filters []Filters `json:"filters"`
	OrderBy string    `json:"order_by"`
	Limit   int       `json:"limit"`
	Offset  int       `json:"offset"`
}
type Filters struct {
	Model    string `json:"model"`
	Property string `json:"property"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type QueryResponse struct {
	ModelName string   `json:"model"`
	Limit     int      `json:"limit"`
	Offset    int      `json:"offset"`
	Records   []Record `json:"filters"`
}

type AutocompleteRequestBody struct {
	Model    string    `json:"model"`
	Property string    `json:"property"`
	Text     string    `json:"text"`
	Filters  []Filters `json:"filters"`
}

type AutocompleteResponse struct {
	ModelName string   `json:"model"`
	Property  string   `json:"property"`
	Values    []string `json:"values"`
}
