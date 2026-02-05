package dotenvgo

import (
	"os"
	"testing"
)

func TestVarUtilities(t *testing.T) {
	key := "TEST_UTIL_VAR"

	t.Run("Set and Unset", func(t *testing.T) {
		Set(key, "util_value")
		if os.Getenv(key) != "util_value" {
			t.Errorf("Set failed")
		}

		Unset(key)
		if _, exists := os.LookupEnv(key); exists {
			t.Errorf("Unset failed")
		}
	})

	t.Run("IsSet", func(t *testing.T) {
		Set(key, "val")
		defer Unset(key)

		v := New[string](key)
		if !v.IsSet() {
			t.Error("Expected IsSet to return true")
		}

		v2 := New[string]("NON_EXISTENT")
		if v2.IsSet() {
			t.Error("Expected IsSet to return false")
		}
	})

	t.Run("Lookup", func(t *testing.T) {
		Set(key, "val")
		defer Unset(key)

		v := New[string](key)
		val, exists := v.Lookup()
		if !exists || val != "val" {
			t.Errorf("Lookup failed: %v, %v", val, exists)
		}

		// Default interaction
		vDef := New[string]("NON_EXISTENT").Default("default")
		val, exists = vDef.Lookup()
		if !exists || val != "default" {
			t.Errorf("Lookup default failed: %v, %v", val, exists)
		}

		// Missing
		vMiss := New[string]("NON_EXISTENT_2")
		_, exists = vMiss.Lookup()
		if exists {
			t.Error("Lookup expected false for missing var")
		}

		// Parser Error
		Set("INT_KEY", "invalid")
		defer Unset("INT_KEY")
		vInt := New[int]("INT_KEY")
		_, exists = vInt.Lookup()
		if exists {
			t.Error("Lookup expected false for invalid value")
		}
	})

	t.Run("MustGet", func(t *testing.T) {
		Set(key, "val")
		defer Unset(key)

		v := New[string](key)
		if val := v.MustGet(); val != "val" {
			t.Errorf("MustGet returned %v", val)
		}
	})
}

func TestExport(t *testing.T) {
	Set("APP_TEST_1", "v1")
	Set("APP_TEST_2", "v2")
	Set("OTHER_VAR", "v3")
	defer Unset("APP_TEST_1")
	defer Unset("APP_TEST_2")
	defer Unset("OTHER_VAR")

	t.Run("Export All", func(t *testing.T) {
		m := Export()
		if m["APP_TEST_1"] != "v1" || m["OTHER_VAR"] != "v3" {
			t.Error("Export failed to return all vars")
		}
	})

	t.Run("Export With Prefix", func(t *testing.T) {
		m := ExportWithPrefix("APP")
		if len(m) < 2 {
			t.Error("ExportWithPrefix returned too few vars")
		}
		if m["APP_TEST_1"] != "v1" {
			t.Error("Missing APP_TEST_1")
		}
		if _, ok := m["OTHER_VAR"]; ok {
			t.Error("Should not include OTHER_VAR")
		}

		// Case insensitivity check if implemented (it is in code)
		m2 := ExportWithPrefix("app")
		if m2["APP_TEST_1"] != "v1" {
			t.Error("ExportWithPrefix (lowercase) failed")
		}
	})
}

func TestLoadDotEnvExtras(t *testing.T) {
	filename := ".env.override"
	content := []byte("TEST_KEY=new_value\n# Comment")
	if err := os.WriteFile(filename, content, 0o644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filename)

	t.Run("Override", func(t *testing.T) {
		Set("TEST_KEY", "old_value")
		defer Unset("TEST_KEY")

		if err := LoadDotEnvOverride(filename); err != nil {
			t.Fatal(err)
		}

		if val := os.Getenv("TEST_KEY"); val != "new_value" {
			t.Errorf("Expected 'new_value', got %q", val)
		}
	})

	t.Run("MustLoadDotEnv", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic unexpected: %v", r)
			}
		}()
		MustLoadDotEnv(filename)
	})

	t.Run("MustLoadDotEnv Panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for missing file")
			}
		}()
		MustLoadDotEnv("non_existent_file.env")
	})
}

// Redefine for this package test
type TestIP struct {
	Value string
}

func (i *TestIP) UnmarshalText(text []byte) error {
	i.Value = string(text)
	return nil
}

func TestGenericFallback(t *testing.T) {
	t.Run("TextUnmarshaler", func(t *testing.T) {
		Set("TEST_IP", "1.1.1.1")
		defer Unset("TEST_IP")

		v := New[TestIP]("TEST_IP")
		val, err := v.GetE()
		if err != nil {
			t.Fatalf("GetE failed: %v", err)
		}
		if val.Value != "1.1.1.1" {
			t.Errorf("Expected 1.1.1.1, got %v", val)
		}
	})

	t.Run("No Parser", func(t *testing.T) {
		Set("NO_PARSER", "val")
		defer Unset("NO_PARSER")

		// struct{} has no parser and no TextUnmarshaler
		v := New[struct{}]("NO_PARSER")
		_, err := v.GetE()
		if err == nil {
			t.Error("Expected error for type with no parser")
		}
	})
}

func TestComplexDotEnv(t *testing.T) {
	filename := ".env.complex"
	content := `
# Comment
SIMPLE=value
QUOTED="quoted value"
SINGLE_QUOTED='single quoted'
WITH_HASH=val#ue
WITH_COMMENT=value # comment
UNCLOSED="unclosed
	`
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filename)

	if err := LoadDotEnvOverride(filename); err != nil {
		t.Fatal(err)
	}

	checks := map[string]string{
		"SIMPLE":        "value",
		"QUOTED":        "quoted value",
		"SINGLE_QUOTED": "single quoted",
		"WITH_HASH":     "val#ue",
		"WITH_COMMENT":  "value",
		"UNCLOSED":      "\"unclosed", // Implementation dependent, usually raw
	}

	for k, expected := range checks {
		if got := os.Getenv(k); got != expected {
			t.Errorf("Key %s: expected %q, got %q", k, expected, got)
		}
	}
}
