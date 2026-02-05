package dotenvgo

import "fmt"

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

// ValidationError is returned when struct validation fails.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("dotenvgo: validation failed for field %q: %s", e.Field, e.Message)
}

// MultiError contains multiple errors from struct loading.
type MultiError struct {
	Errors []error
}

func (e *MultiError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("dotenvgo: %d errors occurred", len(e.Errors))
}

// Unwrap returns the list of errors.
func (e *MultiError) Unwrap() []error {
	return e.Errors
}
