package dotenvgo

import (
	"fmt"
	"strings"
)

// RequiredError is returned when a required environment variable is not set.
type RequiredError struct {
	Key string
}

func (e *RequiredError) Error() string {
	return fmt.Sprintf("dotenvgo: required environment variable %q is not set", e.Key)
}

// ParseError is returned when an environment variable cannot be parsed.
type ParseError struct {
	Key   string
	Value string
	Err   error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("dotenvgo: cannot parse %q=%q: %v", e.Key, e.Value, e.Err)
}

// Unwrap returns the underlying error.
func (e *ParseError) Unwrap() error {
	return e.Err
}

// MultiError contains multiple errors from struct loading.
type MultiError struct {
	Errors []error
}

func (e *MultiError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	// Build detailed error message listing all errors
	var msg strings.Builder
	fmt.Fprintf(&msg, "dotenvgo: %d errors occurred:\n", len(e.Errors))
	for i, err := range e.Errors {
		fmt.Fprintf(&msg, "  [%d] %s\n", i+1, err.Error())
	}
	return msg.String()
}

// Unwrap returns the list of errors.
func (e *MultiError) Unwrap() []error {
	return e.Errors
}
