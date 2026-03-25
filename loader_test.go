package dotenvgo

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

type TestConfig struct {
	Host     string        `env:"HOST" default:"localhost"`
	Port     int           `env:"PORT" default:"8080"`
	Debug    bool          `env:"DEBUG" default:"false"`
	Timeout  time.Duration `env:"TIMEOUT" default:"30s"`
	Required string        `env:"REQUIRED_VAR" required:"true"`
	Alias    string        `env:"ALIAS" default:"fallback"`

	// Slices
	Hosts []string `env:"ALLOWED_HOSTS"`
	IDs   []int    `env:"ALLOWED_IDS" sep:";"`

	// Pointers
	Optional   *int    `env:"OPTIONAL_INT"`
	OptionalDSN *string `env:"OPTIONAL_DSN"`

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
			if cfg.Alias != "fallback" {
				t.Errorf("Expected Alias 'fallback', got %q", cfg.Alias)
			}
			if cfg.Optional != nil {
				t.Errorf("Expected Optional to be nil, got %v", *cfg.Optional)
			}
			if cfg.OptionalDSN != nil {
				t.Errorf("Expected OptionalDSN to be nil, got %q", *cfg.OptionalDSN)
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

	t.Run("Empty Strings Are Loaded", func(t *testing.T) {
		setEnv(t, "ALIAS", "")
		setEnv(t, "OPTIONAL_DSN", "")

		var cfg TestConfig
		err := Load(&cfg)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if cfg.Alias != "" {
			t.Errorf("Expected Alias to be empty string, got %q", cfg.Alias)
		}
		if cfg.OptionalDSN == nil {
			t.Fatal("Expected OptionalDSN to be set")
		}
		if *cfg.OptionalDSN != "" {
			t.Errorf("Expected OptionalDSN to be empty string, got %q", *cfg.OptionalDSN)
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

func TestPointerLeafTypes(t *testing.T) {
	type PointerConfig struct {
		Location *time.Location `env:"LOCATION"`
	}

	setEnv(t, "LOCATION", "UTC")

	var cfg PointerConfig
	if err := Load(&cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Location == nil {
		t.Fatal("Expected Location to be set")
	}
	if cfg.Location.String() != "UTC" {
		t.Errorf("Expected Location UTC, got %q", cfg.Location.String())
	}
}

func TestPointerSlices(t *testing.T) {
	t.Run("Pointer To Slice", func(t *testing.T) {
		type PointerSliceConfig struct {
			Hosts *[]string `env:"HOSTS"`
			Ports *[]int    `env:"PORTS" default:"8080,9090"`
		}

		setEnv(t, "HOSTS", "api,worker")

		var cfg PointerSliceConfig
		if err := Load(&cfg); err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if cfg.Hosts == nil {
			t.Fatal("Expected Hosts to be set")
		}
		if len(*cfg.Hosts) != 2 || (*cfg.Hosts)[0] != "api" || (*cfg.Hosts)[1] != "worker" {
			t.Errorf("Expected Hosts [api worker], got %v", *cfg.Hosts)
		}
		if cfg.Ports == nil {
			t.Fatal("Expected Ports default to be set")
		}
		if len(*cfg.Ports) != 2 || (*cfg.Ports)[0] != 8080 || (*cfg.Ports)[1] != 9090 {
			t.Errorf("Expected Ports [8080 9090], got %v", *cfg.Ports)
		}
	})

	t.Run("Slice Of Pointers", func(t *testing.T) {
		type SlicePointerConfig struct {
			Hosts []*string `env:"HOSTS"`
			IDs   []*int    `env:"IDS" sep:";"`
		}

		setEnv(t, "HOSTS", "api,worker")
		setEnv(t, "IDS", "1; 2;3")

		var cfg SlicePointerConfig
		if err := Load(&cfg); err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if len(cfg.Hosts) != 2 {
			t.Fatalf("Expected 2 Hosts, got %d", len(cfg.Hosts))
		}
		if cfg.Hosts[0] == nil || cfg.Hosts[1] == nil {
			t.Fatal("Expected all Host pointers to be non-nil")
		}
		if *cfg.Hosts[0] != "api" || *cfg.Hosts[1] != "worker" {
			t.Errorf("Expected Hosts [api worker], got [%q %q]", *cfg.Hosts[0], *cfg.Hosts[1])
		}

		if len(cfg.IDs) != 3 {
			t.Fatalf("Expected 3 IDs, got %d", len(cfg.IDs))
		}
		for i, expected := range []int{1, 2, 3} {
			if cfg.IDs[i] == nil {
				t.Fatalf("Expected ID pointer at index %d to be non-nil", i)
			}
			if *cfg.IDs[i] != expected {
				t.Errorf("Expected ID %d at index %d, got %d", expected, i, *cfg.IDs[i])
			}
		}
	})

	t.Run("Pointer To Slice Remains Nil When Missing", func(t *testing.T) {
		type PointerSliceConfig struct {
			Hosts *[]string `env:"HOSTS"`
		}

		var cfg PointerSliceConfig
		if err := Load(&cfg); err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if cfg.Hosts != nil {
			t.Errorf("Expected Hosts to remain nil, got %v", *cfg.Hosts)
		}
	})
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

	t.Run("Required Empty String Does Not Error", func(t *testing.T) {
		setEnv(t, "REQUIRED_VAR", "")

		var cfg TestConfig
		err := Load(&cfg)
		if err != nil {
			t.Fatalf("Expected empty required string to load, got %v", err)
		}
		if cfg.Required != "" {
			t.Errorf("Expected Required to be empty string, got %q", cfg.Required)
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

	t.Run("Nested Errors Are Flattened", func(t *testing.T) {
		type Database struct {
			URL      string `env:"URL" required:"true"`
			User     string `env:"USER" required:"true"`
		}

		type Config struct {
			DB Database `env:"DB"`
		}

		var cfg Config
		err := Load(&cfg)
		if err == nil {
			t.Fatal("Expected error for missing nested required vars")
		}

		var multiErr *MultiError
		if !errors.As(err, &multiErr) {
			t.Fatalf("Expected MultiError, got %T", err)
		}
		if len(multiErr.Errors) != 2 {
			t.Fatalf("Expected 2 flattened errors, got %d", len(multiErr.Errors))
		}

		for _, item := range multiErr.Errors {
			var reqErr *RequiredError
			if !errors.As(item, &reqErr) {
				t.Fatalf("Expected flattened RequiredError, got %T", item)
			}
		}
	})

	t.Run("Slice Pointer Parse Error Keeps Cause", func(t *testing.T) {
		type Config struct {
			IDs []*int `env:"IDS"`
		}

		setEnv(t, "IDS", "1, nope, 3")

		var cfg Config
		err := Load(&cfg)
		if err == nil {
			t.Fatal("Expected parse error for invalid slice element")
		}

		var multiErr *MultiError
		if !errors.As(err, &multiErr) || len(multiErr.Errors) == 0 {
			t.Fatalf("Expected MultiError with errors, got %v", err)
		}

		var parseErr *ParseError
		if !errors.As(multiErr.Errors[0], &parseErr) {
			t.Fatalf("Expected ParseError, got %T", multiErr.Errors[0])
		}
		if !strings.Contains(parseErr.Error(), "cannot parse") {
			t.Fatalf("Expected parse-oriented message, got %q", parseErr.Error())
		}
		if strings.Contains(parseErr.Error(), "no parser registered") {
			t.Fatalf("Expected real parse cause, got %q", parseErr.Error())
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

func TestNestedStructPointers(t *testing.T) {
	type Database struct {
		URL string `env:"URL" default:"localhost"`
	}

	type OptionalDatabase struct {
		URL string `env:"URL"`
	}

	type Config struct {
		DB *Database `env:"DB"`
	}

	t.Run("Nil When No Nested Values Or Defaults Exist", func(t *testing.T) {
		type OptionalConfig struct {
			DB *OptionalDatabase `env:"DB"`
		}

		var cfg OptionalConfig
		if err := Load(&cfg); err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if cfg.DB != nil {
			t.Fatalf("Expected DB to remain nil, got %#v", cfg.DB)
		}
	})

	t.Run("Allocates Pointer When Nested Default Exists", func(t *testing.T) {
		var cfg Config
		if err := Load(&cfg); err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if cfg.DB == nil {
			t.Fatal("Expected DB pointer to be allocated from nested default")
		}
		if cfg.DB.URL != "localhost" {
			t.Errorf("Expected DB.URL 'localhost', got %q", cfg.DB.URL)
		}
	})

	t.Run("Allocates Pointer When Nested Value Exists", func(t *testing.T) {
		setEnv(t, "DB_URL", "postgres://localhost:5432/pointer")

		var cfg Config
		if err := Load(&cfg); err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if cfg.DB == nil {
			t.Fatal("Expected DB pointer to be allocated")
		}
		if cfg.DB.URL != "postgres://localhost:5432/pointer" {
			t.Errorf("Expected DB.URL 'postgres://localhost:5432/pointer', got %q", cfg.DB.URL)
		}
	})

	t.Run("Allocates Pointer With Global Prefix", func(t *testing.T) {
		setEnv(t, "APP_DB_URL", "postgres://localhost:5432/pointer-prefixed")

		var cfg Config
		if err := LoadWithPrefix(&cfg, "APP"); err != nil {
			t.Fatalf("LoadWithPrefix failed: %v", err)
		}
		if cfg.DB == nil {
			t.Fatal("Expected DB pointer to be allocated")
		}
		if cfg.DB.URL != "postgres://localhost:5432/pointer-prefixed" {
			t.Errorf("Expected DB.URL 'postgres://localhost:5432/pointer-prefixed', got %q", cfg.DB.URL)
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
