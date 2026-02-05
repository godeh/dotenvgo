package main

import (
	"fmt"
	"time"

	"github.com/godeh/dotenvgo"
)

func main() {
	fmt.Println("=== Without Environment Variables (using defaults) ===")

	// 1. Define configuration with defaults using generic New[T]
	port := dotenvgo.New[int]("PORT").Default(8080)
	host := dotenvgo.New[string]("HOST").Default("localhost")
	debug := dotenvgo.New[bool]("DEBUG").Default(false)
	timeout := dotenvgo.New[time.Duration]("TIMEOUT").Default(30 * time.Second)
	workers := dotenvgo.New[int]("WORKERS").Default(4)

	fmt.Printf("Port:    %d (default: 8080)\n", port.Get())
	fmt.Printf("Host:    %s (default: localhost)\n", host.Get())
	fmt.Printf("Debug:   %v (default: false)\n", debug.Get())
	fmt.Printf("Timeout: %v (default: 30s)\n", timeout.Get())
	fmt.Printf("Workers: %d (default: 4)\n", workers.Get())

	fmt.Println()
	fmt.Println("=== With Environment Variables (overriding defaults) ===")

	// Simulating environment variables for the sake of example
	dotenvgo.Set("APP_PORT", "3000")
	dotenvgo.Set("APP_DEBUG", "true")
	dotenvgo.Set("APP_TIMEOUT", "1m30s")

	// Same variables but with prefix
	appPort := port.WithPrefix("APP").Get()

	// appHost uses default because APP_HOST is not set
	appHost := host.WithPrefix("APP").Get()
	appDebug := debug.WithPrefix("APP").Get()
	appTimeout := timeout.WithPrefix("APP").Get()

	fmt.Printf("Port:    %d (env: 3000)\n", appPort)
	fmt.Printf("Host:    %s (not set, using default)\n", appHost)
	fmt.Printf("Debug:   %v (env: true)\n", appDebug)
	fmt.Printf("Timeout: %v (env: 1m30s)\n", appTimeout)
}
