package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func main() {
	fmt.Println("🎉 Happy Study - Eino Agent Development")
	fmt.Println()

	ctx := context.Background()

	// Demonstrate Eino's core types: create a message
	msg := schema.UserMessage("Hello, Eino!")
	fmt.Printf("Created message: role=%s, content=%s\n", msg.Role, msg.Content)

	// Create a simple processing function
	processFn := func(ctx context.Context, input map[string]any) (map[string]any, error) {
		result := make(map[string]any)
		for k, v := range input {
			result[k] = v
		}
		result["processed"] = true
		result["eino_version"] = "v0.9.3"
		return result, nil
	}

	// Create a Lambda node from the function
	lambda := compose.InvokableLambda[map[string]any, map[string]any](processFn)

	// Build a chain with the lambda
	chain := compose.NewChain[map[string]any, map[string]any]()
	chain.AppendLambda(lambda)

	// Compile the chain
	runnable, err := chain.Compile(ctx)
	if err != nil {
		fmt.Printf("Chain compilation error: %v\n", err)
		os.Exit(1)
	}

	// Run the chain
	input := map[string]any{"message": "Hello from Happy Study!"}
	output, err := runnable.Invoke(ctx, input)
	if err != nil {
		fmt.Printf("Chain invoke error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Chain output: %+v\n", output)
	fmt.Println()
	fmt.Println("✅ Eino framework is ready for development!")
	fmt.Println()
	fmt.Println("Key packages:")
	fmt.Println("  github.com/cloudwego/eino           - Main framework")
	fmt.Println("  github.com/cloudwego/eino/compose   - Graph/Chain composition")
	fmt.Println("  github.com/cloudwego/eino/schema    - Data types (messages, tools)")
	fmt.Println("  github.com/cloudwego/eino/adk       - Agent Development Kit")
	fmt.Println("  github.com/cloudwego/eino-ext       - Component implementations")
	fmt.Println()
	fmt.Println("📖 Docs: https://www.cloudwego.io/docs/eino/")

	os.Exit(0)
}
