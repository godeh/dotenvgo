package dotenvgo

import (
	"errors"
	"strings"
	"testing"
)

func TestErrors(t *testing.T) {
	t.Run("RequiredError", func(t *testing.T) {
		err := &RequiredError{Key: "TEST_VAR"}
		if err.Error() != "dotenvgo: required environment variable \"TEST_VAR\" is not set" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("ParseError", func(t *testing.T) {
		cause := errors.New("invalid syntax")
		err := &ParseError{Key: "PORT", Value: "abc", Err: cause}

		if err.Unwrap() != cause {
			t.Error("Unwrap did not return cause")
		}

		expected := "dotenvgo: cannot parse \"PORT\"=\"abc\": invalid syntax"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("ValidationError", func(t *testing.T) {
		err := &ValidationError{Field: "Age", Message: "too low"}
		expected := "dotenvgo: validation failed for field \"Age\": too low"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("MultiError", func(t *testing.T) {
		e1 := errors.New("error 1")
		e2 := errors.New("error 2")
		err := &MultiError{Errors: []error{e1, e2}}

		if len(err.Unwrap()) != 2 {
			t.Error("Unwrap returned wrong number of errors")
		}

		// New format: lists all errors
		errMsg := err.Error()
		if !strings.Contains(errMsg, "2 errors occurred") {
			t.Errorf("Expected '2 errors occurred' in message: %s", errMsg)
		}
		if !strings.Contains(errMsg, "[1] error 1") {
			t.Errorf("Expected '[1] error 1' in message: %s", errMsg)
		}
		if !strings.Contains(errMsg, "[2] error 2") {
			t.Errorf("Expected '[2] error 2' in message: %s", errMsg)
		}

		// Single error - returns the error directly
		err = &MultiError{Errors: []error{e1}}
		if err.Error() != "error 1" {
			t.Errorf("Unexpected single error message: %s", err.Error())
		}
	})
}
