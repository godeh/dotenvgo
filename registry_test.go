package dotenvgo

import (
	"reflect"
	"testing"
)

type CustomType int

func TestRegistry(t *testing.T) {
	// 1. Initial State
	typ := reflect.TypeOf(CustomType(0))
	if _, ok := getParser(typ); ok {
		t.Fatal("CustomType should not have a parser yet")
	}

	// 2. Register Parser
	RegisterParser(func(s string) (CustomType, error) {
		if s == "valid" {
			return CustomType(1), nil
		}
		return 0, nil
	})

	// 3. Verify Registration
	parser, ok := getParser(typ)
	if !ok {
		t.Fatal("CustomType should have a parser now")
	}

	// 4. Test Parser
	val, err := parser("valid")
	if err != nil {
		t.Fatalf("Parser failed: %v", err)
	}

	if ct, ok := val.(CustomType); !ok || ct != 1 {
		t.Errorf("Expected CustomType(1), got %v", val)
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Run parallel tests to check for race conditions in registry
	// Note: RegisterParser locks, getParser Rlocks.
	// This test just ensures no panics under basic concurrent load.
	type concurrentType int

	for i := 0; i < 10; i++ {
		t.Run("Concurrent", func(t *testing.T) {
			t.Parallel()

			// Read
			_, _ = getParser(reflect.TypeOf(concurrentType(0)))

			// Write (registering same type repeated not ideal but safe)
			RegisterParser(func(s string) (concurrentType, error) { return 0, nil })
		})
	}
}

func TestRegisteredParsers(t *testing.T) {
	tests := []struct {
		typ      reflect.Type
		input    string
		expected any
	}{
		{reflect.TypeOf(""), "hello", "hello"},
		{reflect.TypeOf(int(0)), "123", 123},
		{reflect.TypeOf(int8(0)), "123", int8(123)},
		{reflect.TypeOf(int16(0)), "123", int16(123)},
		{reflect.TypeOf(int32(0)), "123", int32(123)},
		{reflect.TypeOf(int64(0)), "123", int64(123)},
		{reflect.TypeOf(uint(0)), "123", uint(123)},
		{reflect.TypeOf(uint8(0)), "123", uint8(123)},
		{reflect.TypeOf(uint16(0)), "123", uint16(123)},
		{reflect.TypeOf(uint32(0)), "123", uint32(123)},
		{reflect.TypeOf(uint64(0)), "123", uint64(123)},
		{reflect.TypeOf(float32(0)), "1.5", float32(1.5)},
		{reflect.TypeOf(float64(0)), "1.5", float64(1.5)},
		{reflect.TypeOf(true), "true", true},
		{reflect.TypeOf(true), "1", true},
		{reflect.TypeOf(true), "yes", true},
		{reflect.TypeOf(true), "on", true},
		{reflect.TypeOf(false), "false", false},
		{reflect.TypeOf(false), "0", false},
		{reflect.TypeOf(false), "no", false},
		{reflect.TypeOf(false), "off", false},
		{reflect.TypeOf(false), "invalid", nil}, // Should error
	}

	for _, tc := range tests {
		parser, ok := getParser(tc.typ)
		if !ok {
			t.Errorf("No parser for type %v", tc.typ)
			continue
		}

		val, err := parser(tc.input)
		if tc.expected == nil {
			if err == nil {
				t.Errorf("Expected error for input %q type %v", tc.input, tc.typ)
			}
		} else {
			if err != nil {
				t.Errorf("Parser failed for input %q type %v: %v", tc.input, tc.typ, err)
			}
			if val != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, val)
			}
		}
	}

	// Test Slices
	sliceTyp := reflect.TypeOf([]string{})
	parser, _ := getParser(sliceTyp)

	v, _ := parser("a,b, c")
	s := v.([]string)
	if len(s) != 3 || s[0] != "a" || s[2] != "c" {
		t.Errorf("Slice parser failed: %v", s)
	}

	v, _ = parser("")
	s = v.([]string)
	if len(s) != 0 {
		t.Errorf("Empty slice parser failed")
	}
}
