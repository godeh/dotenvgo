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
	return DefaultLoader.LoadWithPrefix(cfg, "")
}

// LoadWithPrefix populates a struct with a prefix for all env vars.
// For example, LoadWithPrefix(cfg, "APP") will look for APP_PORT instead of PORT.
func LoadWithPrefix(cfg any, prefix string) error {
	return DefaultLoader.LoadWithPrefix(cfg, prefix)
}

// Load populates a struct from environment variables using struct tags.
func (l *Loader) Load(cfg any) error {
	return l.LoadWithPrefix(cfg, "")
}

// LoadWithPrefix populates a struct with a prefix for all env vars.
func (l *Loader) LoadWithPrefix(cfg any, prefix string) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("dotenvgo: cfg must be a non-nil pointer to a struct")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("dotenvgo: cfg must be a pointer to a struct")
	}

	_, err := l.loadStruct(v, prefix)
	return err
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

func (l *Loader) loadStruct(v reflect.Value, prefix string) (bool, error) {
	t := v.Type()
	var errors []error
	loadedAny := false

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		envKey := field.Tag.Get("env")

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		if nestedType, ok := l.nestedStructType(field.Type); ok {
			nestedPrefix := prefix
			if envKey != "" {
				nestedPrefix = joinEnvKey(prefix, envKey)
			}

			if fieldValue.Kind() == reflect.Pointer {
				nestedValue := reflect.New(nestedType).Elem()
				nestedLoaded, err := l.loadStruct(nestedValue, nestedPrefix)
				if err != nil {
					errors = append(errors, err)
				}
				if nestedLoaded {
					target := reflect.New(nestedType)
					target.Elem().Set(nestedValue)
					fieldValue.Set(target)
					loadedAny = true
				}
				continue
			}

			nestedLoaded, err := l.loadStruct(fieldValue, nestedPrefix)
			if err != nil {
				errors = append(errors, err)
			}
			if nestedLoaded {
				loadedAny = true
			}
			continue
		}

		// Handle structs (embedded or named) that don't have a parser/unmarshaler
		if field.Type.Kind() == reflect.Struct {
			// Check if it's a "leaf" type (has parser or implements TextUnmarshaler)
			_, hasParser := l.getParser(field.Type)
			isUnmarshaler := field.Type.Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) ||
				reflect.PointerTo(field.Type).Implements(reflect.TypeFor[encoding.TextUnmarshaler]())

			if !hasParser && !isUnmarshaler {
				continue
			}
		}

		// Get struct tags
		if envKey == "" {
			continue
		}

		defaultValue := field.Tag.Get("default")
		required := field.Tag.Get("required") == "true"

		// Build full key with prefix
		fullKey := joinEnvKey(prefix, envKey)

		// Get value from environment
		rawValue, exists := os.LookupEnv(fullKey)
		value := os.ExpandEnv(rawValue)
		if !exists {
			if required {
				errors = append(errors, &RequiredError{Key: fullKey})
				continue
			}
			value = os.ExpandEnv(defaultValue)
		}

		if !exists && value == "" {
			continue
		}

		// Parse and set value
		if err := l.setField(fieldValue, field.Tag, value); err != nil {
			errors = append(errors, &ParseError{Key: fullKey, Value: value, Err: err})
			continue
		}
		loadedAny = true
	}

	if len(errors) > 0 {
		return loadedAny, &MultiError{Errors: errors}
	}
	return loadedAny, nil
}

func joinEnvKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "_" + key
}

func (l *Loader) nestedStructType(t reflect.Type) (reflect.Type, bool) {
	baseType := t
	if baseType.Kind() == reflect.Pointer {
		baseType = baseType.Elem()
	}

	if baseType.Kind() != reflect.Struct {
		return nil, false
	}
	if l.isLeafType(t) || l.isLeafType(baseType) {
		return nil, false
	}

	return baseType, true
}

func (l *Loader) isLeafType(t reflect.Type) bool {
	textUnmarshaler := reflect.TypeFor[encoding.TextUnmarshaler]()

	if _, ok := l.getParser(t); ok {
		return true
	}
	if t.Implements(textUnmarshaler) {
		return true
	}
	if t.Kind() != reflect.Pointer && reflect.PointerTo(t).Implements(textUnmarshaler) {
		return true
	}

	return false
}

func (l *Loader) setField(field reflect.Value, tag reflect.StructTag, value string) error {
	if field.Kind() == reflect.Pointer {
		if parser, ok := l.getParser(field.Type()); ok {
			parsed, err := parser(value)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(parsed))
			return nil
		}

		target := reflect.New(field.Type().Elem())
		if u, ok := target.Interface().(encoding.TextUnmarshaler); ok {
			if err := u.UnmarshalText([]byte(value)); err != nil {
				return err
			}
			field.Set(target)
			return nil
		}

		if err := l.setField(target.Elem(), tag, value); err != nil {
			return err
		}
		field.Set(target)
		return nil
	}

	// 0. Handle custom separator for slices
	if field.Kind() == reflect.Slice {
		sep := tag.Get("sep")
		if sep != "" || field.Type().Elem().Kind() == reflect.Pointer {
			if sep == "" {
				sep = ","
			}
			slice, err := l.parseSlice(field.Type(), tag, value, sep)
			if err != nil {
				return err
			}
			field.Set(slice)
			return nil
		}
	}

	// 1. Check if type has a registered parser
	if parser, ok := l.getParser(field.Type()); ok {
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

func (l *Loader) parseSlice(sliceType reflect.Type, tag reflect.StructTag, value, sep string) (reflect.Value, error) {
	parts := strings.Split(value, sep)
	slice := reflect.MakeSlice(sliceType, 0, len(parts))
	elemType := sliceType.Elem()

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		elem := reflect.New(elemType).Elem()
		if err := l.setField(elem, tag, part); err != nil {
			return reflect.Value{}, fmt.Errorf("dotenvgo: no parser registered for slice element type %v: %w", elemType, err)
		}
		slice = reflect.Append(slice, elem)
	}

	return slice, nil
}
