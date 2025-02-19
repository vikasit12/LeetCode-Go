package main

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"go/ast"
	"log"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// Limit function code to avoid exceeding GPT-4 token limit
func limitFunctionCode(code string) string {
	lines := strings.Split(code, "\n")
	if len(lines) > 50 {
		lines = lines[:50] // Keep only first 50 lines
		lines = append(lines, "// Function truncated to reduce token usage...")
	}
	return strings.Join(lines, "\n")
}

// Extract modified function names only
func extractModifiedFunctionNames(diffContent string) []string {
	var modifiedFunctions []string
	lines := strings.Split(diffContent, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "+func ") || strings.HasPrefix(line, "-func ") {
			// Extract function name
			parts := strings.Fields(line)
			if len(parts) > 1 {
				functionName := strings.TrimSuffix(parts[1], "(") // Remove '(' from function signature
				modifiedFunctions = append(modifiedFunctions, functionName)
			}
		}
	}
	return modifiedFunctions
}

// Call GPT-4 for PR analysis but only using function names, NOT full diff
func getGPT4Analysis(functionNames []string) string {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	// Join function names into a single message to avoid excessive tokens
	functionList := strings.Join(functionNames, ", ")

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: "You are a Golang expert. Analyze the following modified functions and identify which need new unit tests.",
				},
				{
					Role:    "user",
					Content: fmt.Sprintf("Modified functions in PR: %s", functionList),
				},
			},
		},
	)
	if err != nil {
		log.Fatalf("Error calling OpenAI API: %v", err)
	}

	return resp.Choices[0].Message.Content
}

func main() {
	// Read Git diff (stored in a file)
	diffFile := "diff.patch"
	diffContentBytes, err := os.ReadFile(diffFile)
	if err != nil {
		log.Fatalf("Error reading diff file: %v", err)
	}
	diffContent := string(diffContentBytes)

	// Extract function names instead of full diff
	modifiedFunctions := extractModifiedFunctionNames(diffContent)

	// If no functions are found, exit early
	if len(modifiedFunctions) == 0 {
		fmt.Println("No function changes detected. Skipping GPT analysis.")
		return
	}

	// Call GPT-4 with only function names, reducing token usage
	gptAnalysis := getGPT4Analysis(modifiedFunctions)
	fmt.Println("GPT-4 Analysis:\n", gptAnalysis)
}
