package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// Extract function names from a Git diff
func extractModifiedFunctionNames(diffContent string) []string {
	var modifiedFunctions []string
	lines := strings.Split(diffContent, "\n")

	// Regex to match function declarations in Go
	funcRegex := regexp.MustCompile(`^\+?\s*func\s+([a-zA-Z0-9_]+)\(`)

	for _, line := range lines {
		matches := funcRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			functionName := matches[1]
			modifiedFunctions = append(modifiedFunctions, functionName)
		}
	}

	return modifiedFunctions
}

// Call GPT-4 with only function names, not full diff
func getGPT4Analysis(functionNames []string) string {
	if len(functionNames) == 0 {
		fmt.Println("No valid function names extracted. Skipping GPT-4 analysis.")
		return ""
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	// Join function names into a single message
	functionList := strings.Join(functionNames, ", ")

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: "You are a Golang expert. Identify which of the following modified functions require new unit tests.",
				},
				{
					Role:    "user",
					Content: fmt.Sprintf("Modified functions: %s", functionList),
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

	// Extract function names from the diff
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