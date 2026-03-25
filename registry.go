package dotenvgo

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var errorType = reflect.TypeFor[error]()

// Loader manages the configuration loading and parser registry.
type Loader struct {
	mu       sync.RWMutex
	registry map[reflect.Type]func(string) (any, error)
}

// DefaultLoader is the default loader instance used by global functions.
var DefaultLoader = NewLoader()

// NewLoader creates a new Loader with default parsers registered.
func NewLoader() *Loader {
	l := &Loader{
		registry: make(map[reflect.Type]func(string) (any, error)),
	}
	l.registerDefaults()
	return l
}

// registerDefaults registers the standard parsers.
func (l *Loader) registerDefaults() {
	// String
	l.RegisterParser(func(s string) (string, error) { return s, nil })

	// Integers
	l.RegisterParser(func(s string) (int, error) { return strconv.Atoi(s) })
	l.RegisterParser(func(s string) (int8, error) {
		v, err := strconv.ParseInt(s, 10, 8)
		return int8(v), err
	})
	l.RegisterParser(func(s string) (int16, error) {
		v, err := strconv.ParseInt(s, 10, 16)
		return int16(v), err
	})
	l.RegisterParser(func(s string) (int32, error) {
		v, err := strconv.ParseInt(s, 10, 32)
		return int32(v), err
	})
	l.RegisterParser(func(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) })

	// Unsigned Integers
	l.RegisterParser(func(s string) (uint, error) {
		v, err := strconv.ParseUint(s, 10, 64)
		return uint(v), err
	})
	l.RegisterParser(func(s string) (uint8, error) {
		v, err := strconv.ParseUint(s, 10, 8)
		return uint8(v), err
	})
	l.RegisterParser(func(s string) (uint16, error) {
		v, err := strconv.ParseUint(s, 10, 16)
		return uint16(v), err
	})
	l.RegisterParser(func(s string) (uint32, error) {
		v, err := strconv.ParseUint(s, 10, 32)
		return uint32(v), err
	})
	l.RegisterParser(func(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) })

	// Floats
	l.RegisterParser(func(s string) (float32, error) {
		v, err := strconv.ParseFloat(s, 64)
		return float32(v), err
	})
	l.RegisterParser(func(s string) (float64, error) { return strconv.ParseFloat(s, 64) })

	// Bool
	l.RegisterParser(func(s string) (bool, error) {
		switch strings.ToLower(s) {
		case "true", "1", "yes", "on", "y":
			return true, nil
		case "false", "0", "no", "off", "n":
			return false, nil
		default:
			return false, fmt.Errorf("invalid boolean value: %q", s)
		}
	})

	// Time Duration
	l.RegisterParser(time.ParseDuration)

	// Time Location
	l.RegisterParser(time.LoadLocation)

	// NOTE: Slice types ([]int, []bool, etc.) are automatically supported!
	// When you register a parser for type T, []T is automatically handled
	// by getParser() which generates a slice parser dynamically.
	// The only exception is []string which we register explicitly for efficiency.

	// String Slice (explicit for efficiency, avoiding reflection overhead)
	l.RegisterParser(func(s string) ([]string, error) {
		if s == "" {
			return []string{}, nil
		}
		parts := strings.Split(s, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result, nil
	})
}

// RegisterParser registers a custom parser for a specific type T.
// This parser will be used when loading structs with fields of type T.
func RegisterParser[T any](parser func(string) (T, error)) {
	DefaultLoader.RegisterParser(parser)
}

// RegisterParser registers a custom parser for a specific type T on this Loader instance.
func (l *Loader) RegisterParser(parser any) {
	// We use reflection to get the function type and the return type T
	v := reflect.ValueOf(parser)
	if v.Kind() != reflect.Func {
		panic("parser must be a function")
	}
	t := v.Type()
	if t.NumIn() != 1 || t.In(0).Kind() != reflect.String {
		panic("parser must take a single string argument")
	}
	if t.NumOut() != 2 || !t.Out(1).Implements(errorType) {
		panic("parser must return (T, error)")
	}

	targetType := t.Out(0)

	l.mu.Lock()
	defer l.mu.Unlock()

	l.registry[targetType] = func(s string) (any, error) {
		res := v.Call([]reflect.Value{reflect.ValueOf(s)})
		errResult := res[1]
		switch errResult.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
			if errResult.IsNil() {
				return res[0].Interface(), nil
			}
		}

		errVal := errResult.Interface()
		if errVal != nil {
			return nil, errVal.(error)
		}
		return res[0].Interface(), nil
	}
}

// WithLoader creates a new environment variable of type T using the specified Loader.
// This provides a more fluent API for creating variables with isolated loaders.
//
// Example:
//
//	loader := dotenvgo.NewLoader()
//	loader.RegisterParser(func(s string) (MyType, error) { ... })
//	value := dotenvgo.WithLoader[MyType](loader, "MY_VAR").Get()
func WithLoader[T any](l *Loader, key string) *Var[T] {
	return NewVar[T](l, key)
}

// getParser returns a registered parser for the given type, if one exists.
// If the type is a slice and no direct parser exists, it will automatically
// generate a slice parser if a parser for the element type is registered.
func (l *Loader) getParser(t reflect.Type) (func(string) (any, error), bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// 1. Check for direct parser
	if parser, ok := l.registry[t]; ok {
		return parser, true
	}

	// 2. If it's a slice, try to generate parser from element type
	if t.Kind() == reflect.Slice {
		elemType := t.Elem()
		if elemParser, ok := l.registry[elemType]; ok {
			// Generate slice parser dynamically
			sliceParser := func(s string) (any, error) {
				slice, err := parseSliceValue(t, s, ",", func(part string) (reflect.Value, error) {
					val, err := elemParser(part)
					if err != nil {
						return reflect.Value{}, err
					}
					return reflect.ValueOf(val), nil
				})
				if err != nil {
					return nil, err
				}
				return slice.Interface(), nil
			}
			return sliceParser, true
		}
	}

	return nil, false
}
