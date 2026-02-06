package dotenvgo

import (
	"encoding"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Var represents an environment variable with type-safe access.
type Var[T any] struct {
	key          string
	defaultValue *T
	required     bool
	parser       func(string) (T, error)
	prefix       string
}

// New creates a new environment variable of type T using the default loader.
func New[T any](key string) *Var[T] {
	return NewVar[T](DefaultLoader, key)
}

// NewVar creates a new environment variable of type T using the specified Loader.
// It searches for a registered parser or uses encoding.TextUnmarshaler.
func NewVar[T any](l *Loader, key string) *Var[T] {
	var zero T
	typ := reflect.TypeOf(zero)

	// 1. Check registry
	if p, ok := l.getParser(typ); ok {
		return &Var[T]{
			key: key,
			parser: func(s string) (T, error) {
				v, err := p(s)
				if err != nil {
					return zero, err
				}
				return v.(T), nil
			},
		}
	}

	// 2. Check TextUnmarshaler
	// Better approach: Generic Parser Factory
	return &Var[T]{
		key: key,
		parser: func(s string) (T, error) {
			// Re-lookup registry (fast)
			if p, ok := l.getParser(typ); ok {
				v, err := p(s)
				if err != nil {
					return zero, err
				}
				return v.(T), nil
			}

			// TextUnmarshaler
			valPtr := reflect.New(typ)
			if u, ok := valPtr.Interface().(encoding.TextUnmarshaler); ok {
				if err := u.UnmarshalText([]byte(s)); err != nil {
					return zero, err
				}
				return valPtr.Elem().Interface().(T), nil
			}

			return zero, fmt.Errorf("dotenvgo: no parser registered for type %v", typ)
		},
	}
}

// Default sets the default value if the environment variable is not set.
func (v *Var[T]) Default(value T) *Var[T] {
	v.defaultValue = &value
	return v
}

// Required marks the environment variable as required.
// Get() will panic if the variable is not set.
// GetE() will return an error.
func (v *Var[T]) Required() *Var[T] {
	v.required = true
	return v
}

// WithPrefix adds a prefix to the environment variable key.
// For example, WithPrefix("APP").String("PORT") will look for "APP_PORT".
func (v *Var[T]) WithPrefix(prefix string) *Var[T] {
	v.prefix = prefix
	return v
}

// fullKey returns the full environment variable key with prefix.
func (v *Var[T]) fullKey() string {
	if v.prefix != "" {
		return v.prefix + "_" + v.key
	}
	return v.key
}

// Get returns the value of the environment variable.
// Panics if the variable is required but not set.
func (v *Var[T]) Get() T {
	value, err := v.GetE()
	if err != nil {
		panic(err)
	}
	return value
}

// GetE returns the value of the environment variable or an error.
func (v *Var[T]) GetE() (T, error) {
	var zero T
	key := v.fullKey()
	raw := os.Getenv(key)

	if raw == "" {
		if v.required {
			return zero, &RequiredError{Key: key}
		}
		if v.defaultValue != nil {
			return *v.defaultValue, nil
		}
		return zero, nil
	}

	value, err := v.parser(os.ExpandEnv(raw))
	if err != nil {
		return zero, &ParseError{Key: key, Value: raw, Err: err}
	}

	return value, nil
}

// Lookup returns the value and whether it was set.
func (v *Var[T]) Lookup() (T, bool) {
	var zero T
	key := v.fullKey()
	raw, exists := os.LookupEnv(key)

	if !exists || raw == "" {
		if v.defaultValue != nil {
			return *v.defaultValue, true
		}
		return zero, false
	}

	value, err := v.parser(os.ExpandEnv(raw))
	if err != nil {
		return zero, false
	}

	return value, true
}

// MustGet returns the value or panics if there's an error.
// Alias for Get().
func (v *Var[T]) MustGet() T {
	return v.Get()
}

// IsSet returns whether the environment variable is set.
func (v *Var[T]) IsSet() bool {
	key := v.fullKey()
	_, exists := os.LookupEnv(key)
	return exists
}

// LoadDotEnv loads environment variables from a .env file.
// By default, it does NOT override existing environment variables.
// Pass true as the second argument to override existing variables.
//
// Examples:
//
//	LoadDotEnv(".env")        // doesn't override existing vars
//	LoadDotEnv(".env", false) // same as above
//	LoadDotEnv(".env", true)  // overrides existing vars
func LoadDotEnv(path string, override ...bool) error {
	shouldOverride := false
	if len(override) > 0 {
		shouldOverride = override[0]
	}
	return loadDotEnvInternal(path, shouldOverride)
}

// LoadDotEnvOverride loads environment variables from a .env file.
// It DOES override existing environment variables.
//
// Deprecated: Use LoadDotEnv(path, true) instead.
func LoadDotEnvOverride(path string) error {
	return LoadDotEnv(path, true)
}

func loadDotEnvInternal(path string, override bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Find the first '='
		before, after, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key := strings.TrimSpace(before)
		valPart := strings.TrimSpace(after)
		var value string

		if len(valPart) > 0 {
			quote := valPart[0]
			if quote == '"' || quote == '\'' {
				// Quoted value: look for matching close quote
				// We start looking from index 1
				if endIdx := strings.IndexByte(valPart[1:], quote); endIdx != -1 {
					// endIdx is relative to valPart[1:], so actual index in valPart is endIdx + 1
					value = valPart[1 : endIdx+1]
				} else {
					// Unclosed quote, take valid part or whole?
					// Usually behave as if unquoted or error.
					// For simplicity/robustness, treat as unquoted if not closed properly
					// or just take the whole thing.
					// Let's assume the user meant it to be the value if unclosed.
					value = valPart
				}
			} else {
				// Unquoted value: stop at first '#' if it is preceded by a space
				value = valPart
				for i := 0; i < len(valPart); i++ {
					if valPart[i] == '#' {
						// Comment matches if it's the first char (handled mostly by loop skip, but technically possible here if valPart is just #)
						// OR if preceded by whitespace
						if i == 0 {
							value = ""
							break
						}
						// Check previous char for whitespace
						if i > 0 && (valPart[i-1] == ' ' || valPart[i-1] == '\t') {
							value = strings.TrimSpace(valPart[:i])
							break
						}
					}
				}
			}
		}

		// Only set if not already set (unless override)
		if override || os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}

	return nil
}

// MustLoadDotEnv loads a .env file or panics.
func MustLoadDotEnv(path string) {
	if err := LoadDotEnv(path); err != nil {
		panic(err)
	}
}

// Export returns all environment variables as a map.
func Export() map[string]string {
	result := make(map[string]string)
	for _, env := range os.Environ() {
		idx := strings.Index(env, "=")
		if idx != -1 {
			result[env[:idx]] = env[idx+1:]
		}
	}
	return result
}

// ExportWithPrefix returns environment variables matching a prefix.
func ExportWithPrefix(prefix string) map[string]string {
	result := make(map[string]string)
	prefixUpper := strings.ToUpper(prefix)
	if !strings.HasSuffix(prefixUpper, "_") {
		prefixUpper += "_"
	}

	for _, env := range os.Environ() {
		idx := strings.Index(env, "=")
		if idx != -1 {
			key := env[:idx]
			if strings.HasPrefix(strings.ToUpper(key), prefixUpper) {
				result[key] = env[idx+1:]
			}
		}
	}
	return result
}

// Set sets an environment variable.
func Set(key, value string) {
	_ = os.Setenv(key, value)
}

// Unset removes an environment variable.
func Unset(key string) {
	_ = os.Unsetenv(key)
}
