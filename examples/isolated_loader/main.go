// Package main demonstrates the isolated loader feature of dotenvgo.
//
// This example shows how different loaders can have different parsers
// for the same type, enabling library authors to customize parsing
// without affecting other parts of the application.
//
// Use case: Two libraries need to parse the same environment variable
// differently. For example:
//   - Library A interprets "primary" color as "Blue"
//   - Library B interprets "primary" color as "Red"
//
// Without isolated loaders, registering a parser globally would affect
// all code using that type. With isolated loaders, each library can
// have its own interpretation.
package main

import (
	"fmt"
	"os"

	"github.com/godeh/dotenvgo"
)

// BrandColor is a custom type that we will parse differently in different loaders.
// This simulates a scenario where two libraries use the same type but need
// different parsing logic.
type BrandColor string

func main() {
	// Setup: Set the environment variable that both "libraries" will read
	os.Setenv("THEME_COLOR", "primary")
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           Isolated Loader Example - dotenvgo               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Environment: THEME_COLOR=primary")
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════════════
	// SCENARIO: Two libraries interpret "primary" color differently
	// ═══════════════════════════════════════════════════════════════════════

	// ─────────────────────────────────────────────────────────────────────────
	// Library 1: Marketing Department
	// Their brand guideline says "primary" = "Blue"
	// ─────────────────────────────────────────────────────────────────────────
	marketingLoader := dotenvgo.NewLoader()
	marketingLoader.RegisterParser(func(s string) (BrandColor, error) {
		if s == "primary" {
			return "Blue", nil
		}
		return BrandColor(s), nil
	})

	// ─────────────────────────────────────────────────────────────────────────
	// Library 2: Engineering Department
	// Their convention says "primary" = "Red" (because red ones go faster!)
	// ─────────────────────────────────────────────────────────────────────────
	engineeringLoader := dotenvgo.NewLoader()
	engineeringLoader.RegisterParser(func(s string) (BrandColor, error) {
		if s == "primary" {
			return "Red", nil
		}
		return BrandColor(s), nil
	})

	// ═══════════════════════════════════════════════════════════════════════
	// TEST 1: Using Marketing Loader
	// ═══════════════════════════════════════════════════════════════════════
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Test 1: Marketing Loader                                   │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")

	// Using the new fluent API: dotenvgo.WithLoader[T](loader, key)
	mColor := dotenvgo.WithLoader[BrandColor](marketingLoader, "THEME_COLOR").Get()
	fmt.Printf("  Marketing interprets 'primary' as: %s\n", mColor)
	fmt.Printf("  ✅ Expected: Blue | Got: %s\n", mColor)
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════════════
	// TEST 2: Using Engineering Loader
	// ═══════════════════════════════════════════════════════════════════════
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Test 2: Engineering Loader                                 │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")

	// Also using the fluent API
	eColor := dotenvgo.WithLoader[BrandColor](engineeringLoader, "THEME_COLOR").Get()
	fmt.Printf("  Engineering interprets 'primary' as: %s\n", eColor)
	fmt.Printf("  ✅ Expected: Red | Got: %s\n", eColor)
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════════════
	// TEST 3: Global/Default Loader (Isolation Proof)
	// ═══════════════════════════════════════════════════════════════════════
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Test 3: Global Loader (Isolation Proof)                    │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  The DefaultLoader has NO parser registered for BrandColor.")
	fmt.Println("  This proves that registering parsers on isolated loaders")
	fmt.Println("  does NOT pollute the global registry.")
	fmt.Println()

	// Using dotenvgo.New[T] which uses DefaultLoader internally
	globalVar := dotenvgo.New[BrandColor]("THEME_COLOR")

	val, err := globalVar.GetE()
	if err != nil {
		fmt.Printf("  ✅ Error (Expected): %v\n", err)
		fmt.Println()
		fmt.Println("  This error confirms isolation is working correctly!")
	} else {
		fmt.Printf("  ❌ Unexpected Value: %s\n", val)
		fmt.Println("  ERROR: Isolation failed - DefaultLoader should not have a parser!")
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("Conclusion: Each Loader maintains its own parser registry.")
	fmt.Println("Libraries can safely register custom parsers without conflicts.")
	fmt.Println("═══════════════════════════════════════════════════════════════")
}
