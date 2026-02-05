package main

import (
	"fmt"
	"os"

	"github.com/godeh/dotenvgo"
)

func main() {
	// Create a temporary .env file for demonstration
	envContent := `
# Advanced .env features
APP_NAME='My App'           # Value in single quotes
API_KEY="secret value"      # Value in double quotes
DB_HOST=localhost           # Inline comment
FLAGS=1#2#3                 # Hash inside value (preserved because no space before)
FLAG="1#23"                 # Hash inside value (preserved because no space before)
mixed_quotes='He said "Hello"'
`
	if err := os.WriteFile(".env.example", []byte(envContent), 0644); err != nil {
		panic(err)
	}
	defer os.Remove(".env.example")

	// Load the .env file
	if err := dotenvgo.LoadDotEnvOverride(".env.example"); err != nil {
		panic(err)
	}

	// Print parsed values
	fmt.Printf("APP_NAME: %s\n", os.Getenv("APP_NAME"))
	fmt.Printf("API_KEY: %s\n", os.Getenv("API_KEY"))
	fmt.Printf("DB_HOST: %s\n", os.Getenv("DB_HOST"))
	fmt.Printf("FLAGS: %s\n", os.Getenv("FLAGS"))
	fmt.Printf("FLAG: %s\n", os.Getenv("FLAG"))
	fmt.Printf("mixed_quotes: %s\n", os.Getenv("mixed_quotes"))
}
