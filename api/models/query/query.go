package query

import "github.com/pennsieve/model-service-serverless/api/models"

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
	ModelName string          `json:"model"`
	Limit     int             `json:"limit"`
	Offset    int             `json:"offset"`
	Total     int             `json:"total"s`
	Records   []models.Record `json:"records"`
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

type AutoCompleteParams struct {
	Text     string
	PropName string
}

type FormatParams struct {
	ResultType         FormatType
	AutoCompleteParams AutoCompleteParams
}

type FormatType int64

const (
	RESULTS FormatType = iota
	COUNT
	AUTOCOMPLETE
)

func (q FormatType) String() string {
	switch q {
	case RESULTS:
		return "RESULTS"
	case COUNT:
		return "COUNT"
	case AUTOCOMPLETE:
		return "AUTOCOMPLETE"
	}
	return "UNKNOWN"
}
