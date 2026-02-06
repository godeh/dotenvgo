# dotenvgo

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Zero Dependencies](https://img.shields.io/badge/Dependencies-Zero-green?style=flat-square)](go.mod)
[![Coverage](https://img.shields.io/badge/Coverage-91%25-brightgreen?style=flat-square)](https://github.com/godeh/dotenvgo)

**Type-safe environment variables for Go with generics, struct tags, and isolated loaders.**

## Installation

```bash
go get github.com/godeh/dotenvgo
```

## Quick Examples

### Type-Safe Variables

```go
import "github.com/godeh/dotenvgo"

// With defaults
port := dotenvgo.New[int]("PORT").Default(8080).Get()
debug := dotenvgo.New[bool]("DEBUG").Default(false).Get()

// Required (panics if missing)
dbURL := dotenvgo.New[string]("DATABASE_URL").Required().Get()

// With error handling
apiKey, err := dotenvgo.New[string]("API_KEY").Required().GetE()

// Check existence
if dotenvgo.New[string]("FEATURE_FLAG").IsSet() {
    // ...
}

// Lookup with existence flag
value, exists := dotenvgo.New[string]("OPTIONAL").Lookup()
```

### Struct Loading

```go
type Config struct {
    Host     string        `env:"HOST" default:"localhost"`
    Port     int           `env:"PORT" default:"8080"`
    Debug    bool          `env:"DEBUG"`
    Timeout  time.Duration `env:"TIMEOUT" default:"30s"`
    Database string        `env:"DATABASE_URL" required:"true"`
    Hosts    []string      `env:"ALLOWED_HOSTS" sep:","`
}

var cfg Config
dotenvgo.Load(&cfg)

// Or with prefix (APP_HOST, APP_PORT, etc.)
dotenvgo.LoadWithPrefix(&cfg, "APP")
```

### Load `.env` Files

```go
dotenvgo.LoadDotEnv(".env")        // Don't override existing vars
dotenvgo.LoadDotEnv(".env", true)  // Override existing vars
dotenvgo.MustLoadDotEnv(".env")    // Panic on error
```

### Custom Parsers

```go
type LogLevel int

const (
    DEBUG LogLevel = iota
    INFO
    ERROR
)

// Register globally
dotenvgo.RegisterParser(func(s string) (LogLevel, error) {
    switch strings.ToLower(s) {
    case "debug": return DEBUG, nil
    case "info":  return INFO, nil
    case "error": return ERROR, nil
    default:      return 0, fmt.Errorf("invalid: %s", s)
    }
})

level := dotenvgo.New[LogLevel]("LOG_LEVEL").Default(INFO).Get()

// Slices are automatically supported!
// When you register a parser for T, []T works automatically
levels := dotenvgo.New[[]LogLevel]("LOG_LEVELS").Get() // "debug,info,error"
```

### Isolated Loaders

Use separate loaders when different parts of your application need different parsing logic for the same type:

```go
// Library A: "primary" = Blue
loaderA := dotenvgo.NewLoader()
loaderA.RegisterParser(func(s string) (Color, error) {
    if s == "primary" { return Blue, nil }
    return Color(s), nil
})

// Library B: "primary" = Red  
loaderB := dotenvgo.NewLoader()
loaderB.RegisterParser(func(s string) (Color, error) {
    if s == "primary" { return Red, nil }
    return Color(s), nil
})

// Each loader has its own registry
colorA := dotenvgo.WithLoader[Color](loaderA, "THEME").Get() // Blue
colorB := dotenvgo.WithLoader[Color](loaderB, "THEME").Get() // Red
```

## Supported Types

| Type | Example Values |
|------|----------------|
| `string` | any text |
| `int`, `int8-64` | `42`, `-100` |
| `uint`, `uint8-64` | `42`, `0` |
| `float32`, `float64` | `3.14`, `-0.5` |
| `bool` | `true`, `false`, `1`, `0`, `yes`, `no`, `on`, `off` |
| `time.Duration` | `30s`, `1h30m`, `500ms` |
| `*time.Location` | `America/New_York`, `UTC` |
| `[]string` | `a,b,c` |
| `[]int`, `[]int8-64` | `1,2,3` |
| `[]uint`, `[]uint8-64` | `1,2,3` |
| `[]float32`, `[]float64` | `1.5,2.5` |
| `[]bool` | `true,false,1,0` |
| Custom | Via `RegisterParser` or `encoding.TextUnmarshaler` |

## Struct Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `env` | Variable name | `env:"PORT"` |
| `default` | Default value | `default:"8080"` |
| `required` | Fail if missing | `required:"true"` |
| `sep` | Slice separator | `sep:";"` |

## `.env` File Format

```bash
# Comments
KEY=value
MESSAGE="Hello World"
DEBUG=true # inline comment

# Variable expansion
BASE=/app
CONFIG=${BASE}/config
```

## Error Handling

```go
err := dotenvgo.Load(&cfg)

var reqErr *dotenvgo.RequiredError
if errors.As(err, &reqErr) {
    log.Printf("Missing: %s", reqErr.Key)
}

var multiErr *dotenvgo.MultiError
if errors.As(err, &multiErr) {
    for _, e := range multiErr.Errors {
        log.Println(e)
    }
}
```

## Utilities

```go
dotenvgo.Set("KEY", "value")
dotenvgo.Unset("KEY")

allVars := dotenvgo.Export()
appVars := dotenvgo.ExportWithPrefix("APP")
```

## Examples

See [examples/](./examples) for complete working code:

| Example | Description |
|---------|-------------|
| [basic](./examples/basic) | Simple variable access |
| [struct](./examples/struct) | Struct-based config |
| [file](./examples/file) | Loading `.env` files |
| [expansion](./examples/expansion) | Variable expansion |
| [isolated_loader](./examples/isolated_loader) | Isolated loader demo |

## License

[MIT](LICENSE)
