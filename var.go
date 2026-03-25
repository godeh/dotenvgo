package dotenvgo

import (
	"encoding"
	"fmt"
	"os"
	"reflect"
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

	return &Var[T]{
		key: key,
		parser: func(s string) (T, error) {
			if p, ok := l.getParser(typ); ok {
				v, err := p(s)
				if err != nil {
					return zero, err
				}
				return v.(T), nil
			}

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
	resolved, err := resolveEnvValue(key, v.required)
	if err != nil {
		return zero, err
	}
	if !resolved.exists {
		if v.defaultValue != nil {
			return *v.defaultValue, nil
		}
		return zero, nil
	}

	value, err := v.parser(resolved.value)
	if err != nil {
		return zero, &ParseError{Key: key, Value: resolved.raw, Err: err}
	}

	return value, nil
}

// Lookup returns the value and whether it was set.
func (v *Var[T]) Lookup() (T, bool) {
	var zero T
	key := v.fullKey()
	raw, exists := os.LookupEnv(key)

	if !exists {
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

// fullKey returns the full environment variable key with prefix.
func (v *Var[T]) fullKey() string {
	if v.prefix != "" {
		return v.prefix + "_" + v.key
	}
	return v.key
}
