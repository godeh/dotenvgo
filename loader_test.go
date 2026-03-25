package dotenvgo

import (
	"errors"
	"os"
	"testing"
	"time"
)

type TestConfig struct {
	Host     string        `env:"HOST" default:"localhost"`
	Port     int           `env:"PORT" default:"8080"`
	Debug    bool          `env:"DEBUG" default:"false"`
	Timeout  time.Duration `env:"TIMEOUT" default:"30s"`
	Required string        `env:"REQUIRED_VAR" required:"true"`

	// Slices
	Hosts []string `env:"ALLOWED_HOSTS"`
	IDs   []int    `env:"ALLOWED_IDS" sep:";"`

	// Pointers
	Optional *int `env:"OPTIONAL_INT"`

	// Unexported should be ignored
	secret string `env:"SECRET"`

	// No tag should be ignored
	NoTag string
}

func TestLoad(t *testing.T) {
	// Setup env
	setEnv(t, "REQUIRED_VAR", "required_value")

	t.Run("Defaults", func(t *testing.T) {
		var cfg TestConfig
		err := Load(&cfg)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if cfg.Host != "localhost" {
			t.Errorf("Expected Host 'localhost', got %q", cfg.Host)
		}
		if cfg.Port != 8080 {
			t.Errorf("Expected Port 8080, got %d", cfg.Port)
		}
		if cfg.Debug != false {
			t.Errorf("Expected Debug false, got %v", cfg.Debug)
		}
		if cfg.Timeout != 30*time.Second {
			t.Errorf("Expected Timeout 30s, got %v", cfg.Timeout)
		}
		if cfg.Optional != nil {
			t.Errorf("Expected Optional to be nil, got %v", *cfg.Optional)
		}
		if cfg.secret != "" {
			t.Errorf("Expected unexported field to remain unset, got %q", cfg.secret)
		}
	})

	t.Run("Env Overrides", func(t *testing.T) {
		setEnv(t, "HOST", "127.0.0.1")
		setEnv(t, "PORT", "9090")
		setEnv(t, "DEBUG", "true")
		setEnv(t, "TIMEOUT", "1m")

		var cfg TestConfig
		err := Load(&cfg)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if cfg.Host != "127.0.0.1" {
			t.Errorf("Expected Host '127.0.0.1', got %q", cfg.Host)
		}
		if cfg.Port != 9090 {
			t.Errorf("Expected Port 9090, got %d", cfg.Port)
		}
		if cfg.Debug != true {
			t.Errorf("Expected Debug true, got %v", cfg.Debug)
		}
		if cfg.Timeout != 1*time.Minute {
			t.Errorf("Expected Timeout 1m, got %v", cfg.Timeout)
		}
	})

	t.Run("Pointer Fields", func(t *testing.T) {
		setEnv(t, "OPTIONAL_INT", "42")

		var cfg TestConfig
		err := Load(&cfg)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if cfg.Optional == nil {
			t.Fatal("Expected Optional to be set")
		}
		if *cfg.Optional != 42 {
			t.Errorf("Expected Optional 42, got %d", *cfg.Optional)
		}
	})

	t.Run("Slices", func(t *testing.T) {
		setEnv(t, "ALLOWED_HOSTS", "a,b, c ") // Default comma, trimming
		setEnv(t, "ALLOWED_IDS", "1; 2;3")    // Custom semicolon

		var cfg TestConfig
		err := Load(&cfg)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if len(cfg.Hosts) != 3 || cfg.Hosts[0] != "a" || cfg.Hosts[1] != "b" || cfg.Hosts[2] != "c" {
			t.Errorf("Expected [a, b, c], got %v", cfg.Hosts)
		}
		if len(cfg.IDs) != 3 || cfg.IDs[0] != 1 || cfg.IDs[1] != 2 || cfg.IDs[2] != 3 {
			t.Errorf("Expected [1, 2, 3], got %v", cfg.IDs)
		}
	})
}

func TestPointerDefaults(t *testing.T) {
	type PointerConfig struct {
		OptionalPort *int           `env:"OPTIONAL_PORT" default:"5432"`
		OptionalTTL  *time.Duration `env:"OPTIONAL_TTL" default:"15s"`
	}

	var cfg PointerConfig
	if err := Load(&cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.OptionalPort == nil {
		t.Fatal("Expected OptionalPort to be set from default")
	}
	if *cfg.OptionalPort != 5432 {
		t.Errorf("Expected OptionalPort 5432, got %d", *cfg.OptionalPort)
	}
	if cfg.OptionalTTL == nil {
		t.Fatal("Expected OptionalTTL to be set from default")
	}
	if *cfg.OptionalTTL != 15*time.Second {
		t.Errorf("Expected OptionalTTL 15s, got %v", *cfg.OptionalTTL)
	}
}

func TestLoadWithPrefix(t *testing.T) {
	setEnv(t, "APP_REQUIRED_VAR", "req")
	setEnv(t, "APP_HOST", "app-host")

	var cfg TestConfig
	err := LoadWithPrefix(&cfg, "APP")
	if err != nil {
		t.Fatalf("LoadWithPrefix failed: %v", err)
	}

	if cfg.Host != "app-host" {
		t.Errorf("Expected Host 'app-host', got %q", cfg.Host)
	}
	if cfg.Required != "req" {
		t.Errorf("Expected Required 'req', got %q", cfg.Required)
	}
}

func TestLoadErrors(t *testing.T) {
	t.Run("Nil Pointer", func(t *testing.T) {
		err := Load(nil)
		if err == nil {
			t.Error("Expected error for nil")
		}
	})

	t.Run("Not Pointer", func(t *testing.T) {
		var cfg TestConfig
		err := Load(cfg)
		if err == nil {
			t.Error("Expected error for non-pointer")
		}
	})

	t.Run("Missing Required", func(t *testing.T) {
		os.Unsetenv("REQUIRED_VAR")
		var cfg TestConfig
		err := Load(&cfg)
		if err == nil {
			t.Error("Expected error for missing required var")
		}

		var multiErr *MultiError
		if !errors.As(err, &multiErr) {
			t.Errorf("Expected MultiError, got %T", err)
		}
	})

	t.Run("Parse Error", func(t *testing.T) {
		setEnv(t, "REQUIRED_VAR", "ok")
		setEnv(t, "PORT", "invalid-int")

		var cfg TestConfig
		err := Load(&cfg)
		if err == nil {
			t.Error("Expected error for parsing failure")
		}

		var multiErr *MultiError
		if !errors.As(err, &multiErr) || len(multiErr.Errors) == 0 {
			t.Fatal("Expected MultiError with at least one error")
		}

		var parseErr *ParseError
		if !errors.As(multiErr.Errors[0], &parseErr) {
			t.Errorf("Expected ParseError, got %T", multiErr.Errors[0])
		}
	})
}

func TestMustLoad(t *testing.T) {
	setEnv(t, "REQUIRED_VAR", "ok")

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustLoad panicked unexpectedly: %v", r)
		}
	}()

	var cfg TestConfig
	MustLoad(&cfg)
}

func TestMustLoadPanic(t *testing.T) {
	os.Unsetenv("REQUIRED_VAR")

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustLoad did not panic")
		}
	}()

	var cfg TestConfig
	MustLoad(&cfg)
}

func TestNestedStructs(t *testing.T) {
	type Database struct {
		URL string `env:"URL" default:"localhost"`
	}

	type AppWithoutPrefix struct {
		Name string `env:"NAME"`
		DB   Database
	}

	type AppWithPrefix struct {
		Name string   `env:"NAME"`
		DB   Database `env:"DB"`
	}

	t.Run("Without Struct Prefix Tag", func(t *testing.T) {
		setEnv(t, "NAME", "MyApp")
		setEnv(t, "URL", "postgres://localhost:5432")

		var app AppWithoutPrefix
		if err := Load(&app); err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if app.Name != "MyApp" {
			t.Errorf("Expected Name 'MyApp', got %q", app.Name)
		}
		if app.DB.URL != "postgres://localhost:5432" {
			t.Errorf("Expected DB.URL 'postgres://localhost:5432', got %q", app.DB.URL)
		}
	})

	t.Run("With Struct Prefix Tag", func(t *testing.T) {
		setEnv(t, "NAME", "MyApp")
		setEnv(t, "DB_URL", "postgres://localhost:5432/prefixed")

		var app AppWithPrefix
		if err := Load(&app); err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if app.Name != "MyApp" {
			t.Errorf("Expected Name 'MyApp', got %q", app.Name)
		}
		if app.DB.URL != "postgres://localhost:5432/prefixed" {
			t.Errorf("Expected DB.URL 'postgres://localhost:5432/prefixed', got %q", app.DB.URL)
		}
	})

	t.Run("With Global Prefix And Struct Prefix Tag", func(t *testing.T) {
		setEnv(t, "APP_NAME", "MyApp")
		setEnv(t, "APP_DB_URL", "postgres://localhost:5432/global-prefixed")

		var app AppWithPrefix
		if err := LoadWithPrefix(&app, "APP"); err != nil {
			t.Fatalf("LoadWithPrefix failed: %v", err)
		}

		if app.Name != "MyApp" {
			t.Errorf("Expected Name 'MyApp', got %q", app.Name)
		}
		if app.DB.URL != "postgres://localhost:5432/global-prefixed" {
			t.Errorf("Expected DB.URL 'postgres://localhost:5432/global-prefixed', got %q", app.DB.URL)
		}
	})
}

// Custom type implementing TextUnmarshaler
type CustomIP struct {
	Value string
}

func (c *CustomIP) UnmarshalText(text []byte) error {
	c.Value = "IP:" + string(text)
	return nil
}

func TestCustomUnmarshaler(t *testing.T) {
	type Config struct {
		IP CustomIP `env:"IP"`
	}

	setEnv(t, "IP", "1.2.3.4")

	var cfg Config
	if err := Load(&cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.IP.Value != "IP:1.2.3.4" {
		t.Errorf("Expected 'IP:1.2.3.4', got %q", cfg.IP.Value)
	}
}

// Helper
func setEnv(t *testing.T, key, value string) {
	_ = os.Setenv(key, value)
	t.Cleanup(func() {
		_ = os.Unsetenv(key)
	})
}

func TestMustLoadWithPrefix(t *testing.T) {
	setEnv(t, "APP_REQUIRED_VAR", "ok")

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustLoadWithPrefix panicked unexpectedly: %v", r)
		}
	}()

	var cfg TestConfig
	MustLoadWithPrefix(&cfg, "APP")
}
