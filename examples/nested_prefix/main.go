// Example: Nested structs using the parent env tag as a prefix
package main

import (
	"fmt"
	"os"

	"github.com/godeh/dotenvgo"
)

type Replica struct {
	URL string `env:"URL" default:"postgres://localhost:5432/replica"`
}

type Database struct {
	URL     string  `env:"URL" default:"postgres://localhost:5432/primary"`
	Replica Replica `env:"REPLICA"`
}

type Config struct {
	Name string   `env:"NAME"`
	DB   Database `env:"DB"`
}

func main() {
	// APP_DB_URL comes from the parent struct tag: DB + URL
	// APP_DB_REPLICA_URL composes prefixes across nested structs.
	os.Setenv("APP_NAME", "nested-demo")
	os.Setenv("APP_DB_URL", "postgres://localhost:5432/app-primary")
	os.Setenv("APP_DB_REPLICA_URL", "postgres://localhost:5432/app-replica")

	var cfg Config
	if err := dotenvgo.LoadWithPrefix(&cfg, "APP"); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Nested Prefix Configuration ===")
	fmt.Printf("Name:        %s\n", cfg.Name)
	fmt.Printf("DB URL:      %s\n", cfg.DB.URL)
	fmt.Printf("Replica URL: %s\n", cfg.DB.Replica.URL)
}
