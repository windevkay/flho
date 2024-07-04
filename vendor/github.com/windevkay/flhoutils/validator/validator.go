package validator

import (
	"regexp"
	"slices"
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	Errors map[string]string
}

// New creates a new instance of the Validator struct.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid checks if the Validator instance has any errors.
// It returns true if there are no errors, otherwise false.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error message to the Validator's Errors map.
// If the given key does not exist in the Errors map, it adds the key-value pair to the map.
// The key is used to identify the error, and the message provides a description of the error.
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check checks if the given condition is false and adds an error to the validator if it is.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// PermittedValue checks if the given value is present in the list of permitted values.
// It returns true if the value is found, otherwise false.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// Matches checks if the given value matches the regular expression pattern.
// It returns true if there is a match, otherwise false.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique checks if the given slice of values contains only unique elements.
// It returns true if all elements are unique, and false otherwise.
func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}
