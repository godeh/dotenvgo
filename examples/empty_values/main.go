// Example: Missing versus empty environment values
package main

import (
	"fmt"
	"os"

	"github.com/godeh/dotenvgo"
)

type Config struct {
	DSN *string `env:"DATABASE_URL"`
}

func main() {
	os.Unsetenv("DATABASE_URL")

	missing := dotenvgo.New[string]("DATABASE_URL").Default("postgres://localhost").Get()
	fmt.Printf("Missing DATABASE_URL uses default: %q\n", missing)

	os.Setenv("DATABASE_URL", "")

	empty := dotenvgo.New[string]("DATABASE_URL").Default("postgres://localhost").Get()
	fmt.Printf("Empty DATABASE_URL stays empty: %q\n", empty)

	required, err := dotenvgo.New[string]("DATABASE_URL").Required().GetE()
	fmt.Printf("Required empty DATABASE_URL succeeds: value=%q err=%v\n", required, err)

	var cfg Config
	if err := dotenvgo.Load(&cfg); err != nil {
		panic(err)
	}

	if cfg.DSN == nil {
		fmt.Println("Struct load produced nil pointer")
		return
	}

	fmt.Printf("Struct load keeps pointer to empty string: %q\n", *cfg.DSN)
}
