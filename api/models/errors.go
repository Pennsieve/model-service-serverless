package models

type UnknownModelError struct {
	Model string
}

func (e *UnknownModelError) Error() string {
	return "Unknown model: " + e.Model
}

type UnsupportedOperatorError struct {
	Operator string
}

func (e *UnsupportedOperatorError) Error() string {
	return "Unsupported Operator: " + e.Operator
}

type UnknownModelPropertyError struct {
	PropName string
}

func (e *UnknownModelPropertyError) Error() string {
	return "Unknown model-property: " + e.PropName
}
