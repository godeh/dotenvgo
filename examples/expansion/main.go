package main

import (
	"fmt"
	"os"

	"github.com/godeh/dotenvgo"
)

func main() {
	// Set base variables
	os.Setenv("HOST", "localhost")
	os.Setenv("PORT", "8080")

	// Set variable that uses expansion
	os.Setenv("SERVICE_URL_v1", "http://${HOST}:${PORT}/api/v1")

	// 1. Using standard getter
	url := dotenvgo.New[string]("SERVICE_URL_v1").Get()
	fmt.Printf("Expanded URL V1: %s\n", url)

	// 2. Using struct loader
	type Config struct {
		Host  string `env:"HOST"`
		Port  int    `env:"PORT"`
		URLV1 string `env:"SERVICE_URL_v1"`
		URLV2 string `env:"SERVICE_URL_v2" default:"http://${HOST}:${PORT}/api/v2"`
	}

	var cfg Config
	if err := dotenvgo.Load(&cfg); err != nil {
		panic(err)
	}

	fmt.Printf("Struct Loaded URL V1: %s\n", cfg.URLV1)
	fmt.Printf("Struct Loaded URL V2: %s\n", cfg.URLV2)
}
