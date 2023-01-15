package models

import "fmt"

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

type EmptyError struct{}

func (e *EmptyError) Error() string {
	return "Cannot use empty name"
}

type NameTooLongError struct {
	Name string
}

func (e *NameTooLongError) Error() string {
	return "Name too long: " + e.Name
}

type ValidationError struct {
	Name string
}

func (e *ValidationError) Error() string {
	return "Name must start with [a-zA-Z] or '_'' and only contain [a-zA-Z0-9_] " + e.Name
}

type ModelNameCountError struct {
	Name string
}

func (e *ModelNameCountError) Error() string {
	return fmt.Sprintf("A model with name %s already exists.", e.Name)
}
