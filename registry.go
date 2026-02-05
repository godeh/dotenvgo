package dotenvgo

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	registryMu sync.RWMutex
	registry   = make(map[reflect.Type]func(string) (any, error))
)

func init() {
	// String
	RegisterParser(func(s string) (string, error) { return s, nil })

	// Integers
	RegisterParser(func(s string) (int, error) { return strconv.Atoi(s) })
	RegisterParser(func(s string) (int8, error) {
		v, err := strconv.ParseInt(s, 10, 8)
		return int8(v), err
	})
	RegisterParser(func(s string) (int16, error) {
		v, err := strconv.ParseInt(s, 10, 16)
		return int16(v), err
	})
	RegisterParser(func(s string) (int32, error) {
		v, err := strconv.ParseInt(s, 10, 32)
		return int32(v), err
	})
	RegisterParser(func(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) })

	// Unsigned Integers
	RegisterParser(func(s string) (uint, error) {
		v, err := strconv.ParseUint(s, 10, 64)
		return uint(v), err
	})
	RegisterParser(func(s string) (uint8, error) {
		v, err := strconv.ParseUint(s, 10, 8)
		return uint8(v), err
	})
	RegisterParser(func(s string) (uint16, error) {
		v, err := strconv.ParseUint(s, 10, 16)
		return uint16(v), err
	})
	RegisterParser(func(s string) (uint32, error) {
		v, err := strconv.ParseUint(s, 10, 32)
		return uint32(v), err
	})
	RegisterParser(func(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) })

	// Floats
	RegisterParser(func(s string) (float32, error) {
		v, err := strconv.ParseFloat(s, 64)
		return float32(v), err
	})
	RegisterParser(func(s string) (float64, error) { return strconv.ParseFloat(s, 64) })

	// Bool
	RegisterParser(func(s string) (bool, error) {
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
	RegisterParser(time.ParseDuration)

	// Time Location
	RegisterParser(time.LoadLocation)

	// String Slice
	RegisterParser(func(s string) ([]string, error) {
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
	registryMu.Lock()
	defer registryMu.Unlock()

	var zero T
	t := reflect.TypeOf(zero)

	registry[t] = func(s string) (any, error) {
		return parser(s)
	}
}

// getParser returns a registered parser for the given type, if one exists.
func getParser(t reflect.Type) (func(string) (any, error), bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	parser, ok := registry[t]
	return parser, ok
}
