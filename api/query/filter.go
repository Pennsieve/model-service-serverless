package query

import "fmt"

type Operator int64
type GraphValue interface {
}

const (
	IS Operator = iota
	IS_NOT
	EQUALS
	NOT_EQUALS
	LESS_THAN
	LESS_THAN_EQUALS
	GREATER_THAN
	GREATER_THAN_EQUALS
	STARTS_WITH
	CONTAINS
)

func (o Operator) String() string {
	switch o {
	case IS:
		return "IS"
	case IS_NOT:
		return "IS_NOT"
	case EQUALS:
		return "EQUALS"
	case NOT_EQUALS:
		return "NOT_EQUALS"
	case LESS_THAN:
		return "LESS_THAN"
	case LESS_THAN_EQUALS:
		return "LESS_THAN_EQUALS"
	case GREATER_THAN:
		return "GREATER_THAN"
	case GREATER_THAN_EQUALS:
		return "GREATER_THAN_EQUALS"
	case STARTS_WITH:
		return "STARTS_WITHIS"
	case CONTAINS:
		return "CONTAINS"
	}
	return "UKNOWN"
}

var StringToOperatorDict = map[string]Operator{
	"IS":                  IS,
	"IS_NOT":              IS_NOT,
	"EQUALS":              EQUALS,
	"NOT_EQUALS":          NOT_EQUALS,
	"LESS_THAN":           LESS_THAN,
	"LESS_THAN_EQUALS":    LESS_THAN_EQUALS,
	"GREATER_THAN":        GREATER_THAN,
	"GREATER_THAN_EQUALS": GREATER_THAN_EQUALS,
	"STARTS_WITH":         STARTS_WITH,
	"CONTAINS":            CONTAINS,
}

type Predicate struct {
	field    string
	op       Operator
	argument GraphValue
}

func (p Predicate) cql(i int, modelName string) string {
	return fmt.Sprintf(" %s[%s] %s %s", modelName, p.predicateField(i), p.op.String(), p.predicateValue(i))
}

func (p Predicate) predicateField(i int) string {
	return fmt.Sprintf("predicate_%d_field", i)
}

func (p Predicate) predicateValue(i int) string {
	return fmt.Sprintf("predicate_%d_value", i)
}
