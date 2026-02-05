// Example: Struct-based configuration loading
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/godeh/dotenvgo"
)

// Config defines your application configuration
type Config struct {
	// Server settings
	Host string `env:"HOST" default:"0.0.0.0"`
	Port int    `env:"PORT" default:"8080"`

	// Feature flags
	Debug   bool `env:"DEBUG" default:"false"`
	Verbose bool `env:"VERBOSE" default:"false"`

	// Timeouts
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" default:"30s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" default:"30s"`

	// Timezone
	Location *time.Location `env:"LOCATION" default:"Europe/London"`

	// Database (required)
	DatabaseURL string `env:"DATABASE_URL" required:"true"`

	// Optional settings
	MaxConnections int      `env:"MAX_CONNECTIONS" default:"100"`
	AllowedOrigins []string `env:"ALLOWED_ORIGINS" default:"*"`
}

func main() {
	// Set some environment variables for demo
	os.Setenv("PORT", "3000")
	os.Setenv("DEBUG", "true")
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost/mydb")
	os.Setenv("READ_TIMEOUT", "1m")
	os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000, https://myapp.com")

	// Load configuration
	var cfg Config
	if err := dotenvgo.Load(&cfg); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Print loaded configuration
	fmt.Println("=== Loaded Configuration ===")
	fmt.Printf("Host:            %s\n", cfg.Host)
	fmt.Printf("Port:            %d\n", cfg.Port)
	fmt.Printf("Debug:           %v\n", cfg.Debug)
	fmt.Printf("Verbose:         %v\n", cfg.Verbose)
	fmt.Printf("Read Timeout:    %v\n", cfg.ReadTimeout)
	fmt.Printf("Write Timeout:   %v\n", cfg.WriteTimeout)
	fmt.Printf("Location:        %v\n", cfg.Location)
	fmt.Printf("Database URL:    %s\n", cfg.DatabaseURL)
	fmt.Printf("Max Connections: %d\n", cfg.MaxConnections)
	fmt.Printf("Allowed Origins: %v\n", cfg.AllowedOrigins)
}
