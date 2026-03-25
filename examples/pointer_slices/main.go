// Example: Pointer to slice and slice of pointers
package main

import (
	"fmt"
	"os"

	"github.com/godeh/dotenvgo"
)

type Config struct {
	Hosts      *[]string `env:"HOSTS"`
	DefaultIDs *[]int    `env:"DEFAULT_IDS" default:"10,20,30"`
	Workers    []*string `env:"WORKERS"`
	Ports      []*int    `env:"PORTS" sep:";"`
}

func main() {
	os.Setenv("HOSTS", "api,worker")
	os.Setenv("WORKERS", "alpha,beta")
	os.Setenv("PORTS", "8080; 9090;10010;57770")

	var cfg Config
	if err := dotenvgo.Load(&cfg); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Pointer Slice Configuration ===")
	hosts := derefStringSlice(cfg.Hosts)
	defaultIDs := derefIntSlice(cfg.DefaultIDs)
	workers := derefStringPointers(cfg.Workers)
	ports := derefIntPointers(cfg.Ports)

	fmt.Printf("Hosts:      %v - %v\n", len(hosts), hosts)
	fmt.Printf("DefaultIDs: %v - %v\n", len(defaultIDs), defaultIDs)
	fmt.Printf("Workers:    %v - %v\n", len(workers), workers)
	fmt.Printf("Ports:      %v - %v\n", len(ports), ports)
}

func derefStringSlice(values *[]string) []string {
	if values == nil {
		return nil
	}
	return *values
}

func derefIntSlice(values *[]int) []int {
	if values == nil {
		return nil
	}
	return *values
}

func derefStringPointers(values []*string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == nil {
			continue
		}
		result = append(result, *value)
	}
	return result
}

func derefIntPointers(values []*int) []int {
	result := make([]int, 0, len(values))
	for _, value := range values {
		if value == nil {
			continue
		}
		result = append(result, *value)
	}
	return result
}
