package main

import (
	"context"
	"fmt"
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// Generate unit tests using GPT-4
func generateTestCases(functionName, functionCode string) string {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: "You are a Golang expert. Generate a high-quality unit test for the given function, ensuring coverage of edge cases and using appropriate mocks where necessary.",
				},
				{
					Role:    "user",
					Content: fmt.Sprintf("Generate a unit test for this function:\n\n%s", functionCode),
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
	functionName := os.Args[1]
	functionCode := os.Args[2] // Extracted function code

	testCode := generateTestCases(functionName, functionCode)

	testFile := fmt.Sprintf("%s_test.go", functionName)
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		log.Fatalf("Error writing test file: %v", err)
	}

	fmt.Println("Generated test case for:", functionName)
}
