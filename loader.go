package dotenvgo

import (
	"encoding"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Load populates a struct from environment variables using struct tags.
//
// Supported tags:
//   - env:"VAR_NAME" - the environment variable name
//   - default:"value" - default value if not set
//   - required:"true" - marks the field as required
//
// Example:
//
//	type Config struct {
//	    Port     int           `env:"PORT" default:"8080"`
//	    Debug    bool          `env:"DEBUG" default:"false"`
//	    Database string        `env:"DATABASE_URL" required:"true"`
//	    Timeout  time.Duration `env:"TIMEOUT" default:"30s"`
//	}
func Load(cfg any) error {
	return LoadWithPrefix(cfg, "")
}

// LoadWithPrefix populates a struct with a prefix for all env vars.
// For example, LoadWithPrefix(cfg, "APP") will look for APP_PORT instead of PORT.
func LoadWithPrefix(cfg any, prefix string) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("dotenvgo: cfg must be a non-nil pointer to a struct")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("dotenvgo: cfg must be a pointer to a struct")
	}

	return loadStruct(v, prefix)
}

// MustLoad is like Load but panics on error.
func MustLoad(cfg any) {
	if err := Load(cfg); err != nil {
		panic(err)
	}
}

// MustLoadWithPrefix is like LoadWithPrefix but panics on error.
func MustLoadWithPrefix(cfg any, prefix string) {
	if err := LoadWithPrefix(cfg, prefix); err != nil {
		panic(err)
	}
}

func loadStruct(v reflect.Value, prefix string) error {
	t := v.Type()
	var errors []error

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Handle structs (embedded or named) that don't have a parser/unmarshaler
		if field.Type.Kind() == reflect.Struct {
			// Check if it's a "leaf" type (has parser or implements TextUnmarshaler)
			_, hasParser := getParser(field.Type)
			isUnmarshaler := field.Type.Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) ||
				reflect.PointerTo(field.Type).Implements(reflect.TypeFor[encoding.TextUnmarshaler]())

			if !hasParser && !isUnmarshaler {
				// Recurse
				if err := loadStruct(fieldValue, prefix); err != nil {
					errors = append(errors, err)
				}
				continue
			}
		}

		// Get struct tags
		envKey := field.Tag.Get("env")
		if envKey == "" {
			continue
		}

		defaultValue := field.Tag.Get("default")
		required := field.Tag.Get("required") == "true"

		// Build full key with prefix
		fullKey := envKey
		if prefix != "" {
			fullKey = prefix + "_" + envKey
		}

		// Get value from environment
		value := os.ExpandEnv(os.Getenv(fullKey))
		if value == "" {
			if required {
				errors = append(errors, &RequiredError{Key: fullKey})
				continue
			}
			value = os.ExpandEnv(defaultValue)
		}

		if value == "" {
			continue
		}

		// Parse and set value
		if err := setField(fieldValue, field.Tag, value); err != nil {
			errors = append(errors, &ParseError{Key: fullKey, Value: value, Err: err})
		}
	}

	if len(errors) > 0 {
		return &MultiError{Errors: errors}
	}
	return nil
}

func setField(field reflect.Value, tag reflect.StructTag, value string) error {
	// 0. Handle custom separator for slices
	if field.Kind() == reflect.Slice {
		sep := tag.Get("sep")
		if sep != "" {
			parts := strings.Split(value, sep)
			slice := reflect.MakeSlice(field.Type(), 0, len(parts))
			elemType := field.Type().Elem()

			// Find parser for element type
			parser, ok := getParser(elemType)
			if !ok {
				// Fallback to TextUnmarshaler for element?
				// For now error if no parser for element
				return fmt.Errorf("dotenvgo: no parser registered for slice element type %v", elemType)
			}

			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				parsed, err := parser(p)
				if err != nil {
					return err
				}
				slice = reflect.Append(slice, reflect.ValueOf(parsed))
			}
			field.Set(slice)
			return nil
		}
	}

	// 1. Check if type has a registered parser
	if parser, ok := getParser(field.Type()); ok {
		parsed, err := parser(value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(parsed))
		return nil
	}

	// 2. Check if field implements encoding.TextUnmarshaler
	if field.CanAddr() {
		// Try pointer receiver
		if u, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return u.UnmarshalText([]byte(value))
		}
	} else if u, ok := field.Interface().(encoding.TextUnmarshaler); ok {
		// Try value receiver (less common for mutation but possible)
		return u.UnmarshalText([]byte(value))
	}

	return fmt.Errorf("dotenvgo: no parser registered for type %v", field.Type())
}
