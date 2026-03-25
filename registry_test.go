package dotenvgo

import (
	"fmt"
	"reflect"
	"testing"
)

type CustomType int

func TestRegistry(t *testing.T) {
	// 1. Initial State
	typ := reflect.TypeOf(CustomType(0))
	if _, ok := DefaultLoader.getParser(typ); ok {
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
	parser, ok := DefaultLoader.getParser(typ)
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

	// 5. Test automatic slice support - []CustomType should work automatically!
	sliceTyp := reflect.TypeFor[[]CustomType]()
	sliceParser, ok := DefaultLoader.getParser(sliceTyp)
	if !ok {
		t.Fatal("[]CustomType should have an auto-generated parser")
	}

	sliceVal, err := sliceParser("valid, valid")
	if err != nil {
		t.Fatalf("Slice parser failed: %v", err)
	}

	customs := sliceVal.([]CustomType)
	if len(customs) != 2 || customs[0] != 1 || customs[1] != 1 {
		t.Errorf("Expected [1, 1], got %v", customs)
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
			_, _ = DefaultLoader.getParser(reflect.TypeOf(concurrentType(0)))

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
		{reflect.TypeFor[string](), "hello", "hello"},
		{reflect.TypeFor[int](), "123", 123},
		{reflect.TypeFor[int8](), "123", int8(123)},
		{reflect.TypeFor[int16](), "123", int16(123)},
		{reflect.TypeFor[int32](), "123", int32(123)},
		{reflect.TypeFor[int64](), "123", int64(123)},
		{reflect.TypeFor[uint](), "123", uint(123)},
		{reflect.TypeFor[uint8](), "123", uint8(123)},
		{reflect.TypeFor[uint16](), "123", uint16(123)},
		{reflect.TypeFor[uint32](), "123", uint32(123)},
		{reflect.TypeFor[uint64](), "123", uint64(123)},
		{reflect.TypeFor[float32](), "1.5", float32(1.5)},
		{reflect.TypeFor[float64](), "1.5", float64(1.5)},
		{reflect.TypeFor[bool](), "true", true},
		{reflect.TypeFor[bool](), "1", true},
		{reflect.TypeFor[bool](), "yes", true},
		{reflect.TypeFor[bool](), "on", true},
		{reflect.TypeFor[bool](), "false", false},
		{reflect.TypeFor[bool](), "0", false},
		{reflect.TypeFor[bool](), "no", false},
		{reflect.TypeFor[bool](), "off", false},
		{reflect.TypeFor[bool](), "invalid", nil}, // Should error
	}

	for _, tc := range tests {
		parser, ok := DefaultLoader.getParser(tc.typ)
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
	sliceTyp := reflect.TypeFor[[]string]()
	parser, _ := DefaultLoader.getParser(sliceTyp)

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

	// Test []int
	intSliceTyp := reflect.TypeFor[[]int]()
	intParser, _ := DefaultLoader.getParser(intSliceTyp)

	vi, err := intParser("1, 2, 3")
	if err != nil {
		t.Errorf("[]int parser failed: %v", err)
	}
	ints := vi.([]int)
	if len(ints) != 3 || ints[0] != 1 || ints[2] != 3 {
		t.Errorf("[]int parser returned wrong values: %v", ints)
	}

	// Test []float64
	floatSliceTyp := reflect.TypeFor[[]float64]()
	floatParser, _ := DefaultLoader.getParser(floatSliceTyp)

	vf, err := floatParser("1.5, 2.5, 3.5")
	if err != nil {
		t.Errorf("[]float64 parser failed: %v", err)
	}
	floats := vf.([]float64)
	if len(floats) != 3 || floats[0] != 1.5 || floats[2] != 3.5 {
		t.Errorf("[]float64 parser returned wrong values: %v", floats)
	}

	// Test []bool
	boolSliceTyp := reflect.TypeFor[[]bool]()
	boolParser, _ := DefaultLoader.getParser(boolSliceTyp)

	vb, err := boolParser("true, false, yes, no, 1, 0")
	if err != nil {
		t.Errorf("[]bool parser failed: %v", err)
	}
	bools := vb.([]bool)
	expected := []bool{true, false, true, false, true, false}
	if len(bools) != len(expected) {
		t.Errorf("[]bool parser returned wrong length: %v", bools)
	}
	for i, b := range bools {
		if b != expected[i] {
			t.Errorf("[]bool parser returned wrong value at index %d: got %v, expected %v", i, b, expected[i])
		}
	}

	// Test []int with invalid value
	_, err = intParser("1, invalid, 3")
	if err == nil {
		t.Error("[]int parser should fail on invalid input")
	}
}

type ValidationStatus int

const (
	StatusUnknown ValidationStatus = iota
	StatusValid
	StatusInvalid
)

func TestLoaderIsolation(t *testing.T) {
	// Create two independent loaders
	l1 := NewLoader()
	l2 := NewLoader()

	// Register parser for l1: "valid" -> StatusValid (1)
	l1.RegisterParser(func(s string) (ValidationStatus, error) {
		if s == "valid" {
			return StatusValid, nil
		}
		return StatusUnknown, nil
	})

	// Register parser for l2: "valid" -> StatusInvalid (2)
	l2.RegisterParser(func(s string) (ValidationStatus, error) {
		if s == "valid" {
			return StatusInvalid, nil
		}
		return StatusUnknown, nil
	})

	// Test l1 behavior
	// We use NewVar generic function passing the loader
	// We need to simulate parsing or use the variable. Since we can't set env var easily in parallel tests without potential race if we used os.Setenv,
	// but NewVar returns a Var with a parser closure. We can test that parser closure if we could access it?
	// No, Var struct fields are unexported.
	// But we can use .Get() or .GetE() which reads from OS.
	// Or we can just inspect if the creation worked if we had a way.

	// Actually, dotenvgo.Var reads from os.Getenv.
	// To test, we need to set an env var.
	// os.Setenv is global.
	// This test should run purely on internal logic if possible, or we accept setting env var.

	t.Setenv("ISOLATION_TEST_VAR", "valid")

	val1 := NewVar[ValidationStatus](l1, "ISOLATION_TEST_VAR").Get()
	if val1 != StatusValid {
		t.Errorf("Loader 1 failed: expected StatusValid (1), got %v", val1)
	}

	val2 := NewVar[ValidationStatus](l2, "ISOLATION_TEST_VAR").Get()
	if val2 != StatusInvalid {
		t.Errorf("Loader 2 failed: expected StatusInvalid (2), got %v", val2)
	}

	// 3. Test l3 (DefaultLoader) - should have NO parser for ValidationStatus
	// This ensures we didn't pollute the global registry
	// Note: New[T] uses DefaultLoader
	v3 := New[ValidationStatus]("ISOLATION_TEST_VAR")
	// Accessing it should error or return default?
	// GetE calls parser. New factory returned a parser that returns error if not found.
	_, err := v3.GetE()
	if err == nil {
		t.Error("DefaultLoader should NOT have a parser for ValidationStatus")
	}
}

type fakeError struct {
	message string
}

func (e *fakeError) Error() string {
	return e.message
}

func TestRegisterParserValidation(t *testing.T) {
	t.Run("accepts custom error type", func(t *testing.T) {
		loader := NewLoader()
		type parserType int

		loader.RegisterParser(func(s string) (parserType, *fakeError) {
			return parserType(len(s)), nil
		})

		parser, ok := loader.getParser(reflect.TypeFor[parserType]())
		if !ok {
			t.Fatal("expected parser to be registered")
		}

		value, err := parser("abc")
		if err != nil {
			t.Fatalf("expected parser to succeed: %v", err)
		}
		if value.(parserType) != 3 {
			t.Fatalf("expected parsed value 3, got %v", value)
		}
	})

	t.Run("rejects non error second return", func(t *testing.T) {
		loader := NewLoader()
		assertPanicsWithMessage(t, "parser must return (T, error)", func() {
			loader.RegisterParser(func(s string) (int, fmt.Stringer) {
				return 0, nil
			})
		})
	})

	t.Run("rejects wrong signatures", func(t *testing.T) {
		loader := NewLoader()

		assertPanicsWithMessage(t, "parser must be a function", func() {
			loader.RegisterParser(123)
		})

		assertPanicsWithMessage(t, "parser must take a single string argument", func() {
			loader.RegisterParser(func(int) (string, error) {
				return "", nil
			})
		})

		assertPanicsWithMessage(t, "parser must return (T, error)", func() {
			loader.RegisterParser(func(string) string {
				return ""
			})
		})
	})
}

func assertPanicsWithMessage(t *testing.T, message string, fn func()) {
	t.Helper()

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("expected panic %q", message)
		}

		panicMessage, ok := recovered.(string)
		if !ok {
			t.Fatalf("expected panic string %q, got %T", message, recovered)
		}

		if panicMessage != message {
			t.Fatalf("expected panic %q, got %q", message, panicMessage)
		}
	}()

	fn()
}
