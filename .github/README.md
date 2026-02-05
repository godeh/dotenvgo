# dotenvgo

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Zero Dependencies](https://img.shields.io/badge/Dependencies-Zero-green?style=flat-square)](go.mod)

**Type-safe, zero-dependency environment variable management for Go.**

`dotenvgo` provides a modern, generic-based API for loading environment variables with full type safety, default values, validation, and struct tag support.

## ✨ Features

- 🔒 **Type-Safe** - Generic `New[T]` API with compile-time type checking
- 📦 **Zero Dependencies** - Only uses Go standard library
- 🏷️ **Struct Tags** - Load complex configs into structs with tags
- 📄 **`.env` File Support** - Load from `.env` files
- 🔌 **Extensible** - Register custom parsers
- ⚡ **Variable Expansion** - Supports `$VAR` and `${VAR}` expansion
- 🔧 **Prefix Support** - Namespace variables with prefixes (e.g., `APP_PORT`)

## 📦 Installation

```bash
go get github.com/godeh/dotenvgo
```

## 🚀 Quick Start

### Basic Usage

```go
import "github.com/godeh/dotenvgo"

// Simple variable access with defaults
port := dotenvgo.New[int]("PORT").Default(8080).Get()
host := dotenvgo.New[string]("HOST").Default("localhost").Get()
debug := dotenvgo.New[bool]("DEBUG").Default(false).Get()

// Required variables (panics if not set)
dbURL := dotenvgo.New[string]("DATABASE_URL").Required().Get()

// Explicit error handling
apiKey, err := dotenvgo.New[string]("API_KEY").Required().GetE()
if err != nil {
    log.Fatal("API_KEY is missing")
}

// Check if variable is set
if dotenvgo.New[string]("OPTIONAL_VAR").IsSet() {
    // Variable exists
}

// Lookup returns value and existence flag
value, exists := dotenvgo.New[string]("MY_VAR").Default("fallback").Lookup()
```

### With Prefix

```go
// Look for APP_PORT instead of PORT
port := dotenvgo.New[int]("PORT").WithPrefix("APP").Default(8080).Get()
```

### Load from `.env` File

```go
// Load without overriding existing env vars
if err := dotenvgo.LoadDotEnv(".env"); err != nil {
    log.Fatal(err)
}

// Load and override existing env vars
dotenvgo.LoadDotEnvOverride(".env")

// Panic on error
dotenvgo.MustLoadDotEnv(".env")
```

## 🏷️ Struct Tags

Load complex configurations into structs using tags.

```go
type Config struct {
    Host     string        `env:"HOST" default:"localhost"`
    Port     int           `env:"PORT" default:"8080"`
    Debug    bool          `env:"DEBUG" default:"false"`
    Timeout  time.Duration `env:"TIMEOUT" default:"30s"`
    Database string        `env:"DATABASE_URL" required:"true"`
    
    // Slice with custom separator
    Hosts    []string      `env:"ALLOWED_HOSTS" sep:";"`
}

var cfg Config

// Load from environment
if err := dotenvgo.Load(&cfg); err != nil {
    log.Fatal(err)
}

// Or with prefix (looks for APP_HOST, APP_PORT, etc.)
if err := dotenvgo.LoadWithPrefix(&cfg, "APP"); err != nil {
    log.Fatal(err)
}

// Panic version
dotenvgo.MustLoad(&cfg)
dotenvgo.MustLoadWithPrefix(&cfg, "APP")
```

### Supported Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `env` | Environment variable name | `env:"PORT"` |
| `default` | Default value if not set | `default:"8080"` |
| `required` | Fails if variable is not set | `required:"true"` |
| `sep` | Custom separator for slices | `sep:";"` |

## 📋 Supported Types

### Primitives
- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `bool` - Accepts: `true/false`, `1/0`, `yes/no`, `on/off`, `y/n`

### Time Types
- `time.Duration` - e.g., `"1h30m"`, `"30s"`, `"500ms"`
- `*time.Location` - e.g., `"America/New_York"`, `"Europe/London"`

### Collections
- `[]string` - Comma-separated by default: `"a,b,c"`

### Custom Types
- Any type implementing `encoding.TextUnmarshaler`
- Custom parsers via `RegisterParser`

## 🔌 Custom Parsers

Register custom parsers for your own types:

```go
type LogLevel int

const (
    LevelDebug LogLevel = iota
    LevelInfo
    LevelWarn
    LevelError
)

func init() {
    dotenvgo.RegisterParser(func(s string) (LogLevel, error) {
        switch strings.ToLower(s) {
        case "debug":
            return LevelDebug, nil
        case "info":
            return LevelInfo, nil
        case "warn":
            return LevelWarn, nil
        case "error":
            return LevelError, nil
        default:
            return 0, fmt.Errorf("invalid log level: %s", s)
        }
    })
}

// Now you can use LogLevel directly
level := dotenvgo.New[LogLevel]("LOG_LEVEL").Default(LevelInfo).Get()
```

## 🛠️ Utility Functions

```go
// Set/Unset environment variables
dotenvgo.Set("KEY", "value")
dotenvgo.Unset("KEY")

// Export all environment variables as map
allVars := dotenvgo.Export()

// Export variables with a specific prefix
appVars := dotenvgo.ExportWithPrefix("APP")
```

## 🔍 Variable Expansion

Environment variable expansion is supported in values:

```go
// .env file
# BASE_PATH=/app
# CONFIG_PATH=${BASE_PATH}/config

// In code, CONFIG_PATH will be "/app/config"
```

## 📄 `.env` File Format

```bash
# Comments start with #
KEY=value

# Quoted values preserve spaces
MESSAGE="Hello World"
NAME='John Doe'

# Inline comments (after space)
DEBUG=true # This is a comment

# Variable expansion
BASE_URL=http://localhost
API_URL=${BASE_URL}/api
```

## ⚠️ Error Handling

The library provides structured error types:

```go
// RequiredError - when a required variable is missing
// ParseError - when a value cannot be parsed to the target type  
// MultiError - when struct loading has multiple errors

if err := dotenvgo.Load(&cfg); err != nil {
    var reqErr *dotenvgo.RequiredError
    if errors.As(err, &reqErr) {
        fmt.Printf("Missing required: %s\n", reqErr.Key)
    }
    
    var multiErr *dotenvgo.MultiError
    if errors.As(err, &multiErr) {
        for _, e := range multiErr.Errors {
            fmt.Println(e)
        }
    }
}
```

## 📂 Examples

See the [examples](./examples) directory for complete working examples:

- [basic](./examples/basic) - Simple variable access with defaults and prefixes
- [struct](./examples/struct) - Struct-based configuration loading
- [file](./examples/file) - Loading from `.env` files
- [expansion](./examples/expansion) - Variable expansion demonstration

## 📄 License

[MIT](LICENSE)
