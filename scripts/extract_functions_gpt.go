package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// Extract modified functions using Go AST
func extractModifiedFunctions(filePath string) []string {
	// Skip vendor directory
	if strings.Contains(filePath, "vendor/") {
		fmt.Println("Skipping vendor file:", filePath)
		return nil
	}
	src, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		log.Fatalf("Error parsing Go file: %v", err)
	}

	var functions []string
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			functions = append(functions, fn.Name.Name)
		}
		return true
	})

	return functions
}

// Call GPT-4 for PR analysis
func getGPT4Analysis(codeDiff string) string {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: "You are a Go language expert. Analyze the following code diff and identify key function changes along with their impact.",
				},
				{
					Role:    "user",
					Content: codeDiff,
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
	filePath := os.Args[1]
	modifiedFunctions := extractModifiedFunctions(filePath)

	fmt.Println("Modified functions:", modifiedFunctions)

	diffFile := "diff.patch"
	diffContent, err := ioutil.ReadFile(diffFile)
	if err != nil {
		log.Fatalf("Error reading diff file: %v", err)
	}

	gptAnalysis := getGPT4Analysis(string(diffContent))
	fmt.Println("GPT-4 Analysis:\n", gptAnalysis)
}
