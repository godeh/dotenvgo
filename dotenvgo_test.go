package dotenvgo

import (
	"os"
	"testing"
)

func TestEnvVar(t *testing.T) {
	// Setup
	key := "TEST_ENV_VAR"
	_ = os.Setenv(key, "test_value")
	defer func() { _ = os.Unsetenv(key) }()

	t.Run("New String", func(t *testing.T) {
		v := New[string](key)
		if val := v.Get(); val != "test_value" {
			t.Errorf("Expected 'test_value', got %v", val)
		}
	})

	t.Run("Default Value", func(t *testing.T) {
		v := New[string]("NON_EXISTENT").Default("default")
		if val := v.Get(); val != "default" {
			t.Errorf("Expected 'default', got %v", val)
		}
	})

	t.Run("Required Var", func(t *testing.T) {
		v := New[string]("NON_EXISTENT").Required()
		_, err := v.GetE()
		if err == nil {
			t.Error("Expected error for missing required var")
		}
	})

	t.Run("With Prefix", func(t *testing.T) {
		_ = os.Setenv("APP_PORT", "8080")
		defer func() { _ = os.Unsetenv("APP_PORT") }()

		v := New[int]("PORT").WithPrefix("APP")
		if val := v.Get(); val != 8080 {
			t.Errorf("Expected 8080, got %v", val)
		}
	})
}

func TestLoadDotEnv(t *testing.T) {
	content := []byte("TEST_KEY=test_value\n# Comment\nQUOTED=\"value with spaces\"")
	filename := ".env.test"
	if err := os.WriteFile(filename, content, 0o644); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(filename) }()
	defer func() { _ = os.Unsetenv("TEST_KEY") }()
	defer func() { _ = os.Unsetenv("QUOTED") }()

	if err := LoadDotEnv(filename); err != nil {
		t.Fatalf("LoadDotEnv failed: %v", err)
	}

	if val := os.Getenv("TEST_KEY"); val != "test_value" {
		t.Errorf("Expected 'test_value', got %q", val)
	}
	if val := os.Getenv("QUOTED"); val != "value with spaces" {
		t.Errorf("Expected 'value with spaces', got %q", val)
	}
}

func TestParsers(t *testing.T) {
	_ = os.Setenv("INT_VAL", "123")
	_ = os.Setenv("BOOL_VAL", "true")
	defer func() { _ = os.Unsetenv("INT_VAL") }()
	defer func() { _ = os.Unsetenv("BOOL_VAL") }()

	if v := New[int]("INT_VAL").Get(); v != 123 {
		t.Errorf("Expected 123, got %v", v)
	}
	if v := New[bool]("BOOL_VAL").Get(); v != true {
		t.Errorf("Expected true, got %v", v)
	}
}

// Mock Generic Parser to avoid reflect error in newGeneric logic if it relies on registry
func init() {
	// Register types used in tests if not implicitly available via dotenvgo logic
	// The current implementation of dotenvgo uses a registry or generic unmarshaler.
	// Since we are black-box testing, we assume int/string/bool support is built-in or registered.
	// Based on dotenvgo.go, it uses `getParser`. We assume `registry.go` (not shown here but referenced in file list) exists.
}
