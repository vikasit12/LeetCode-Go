package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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
	if strings.Contains(filePath, "scripts/") {
		fmt.Println("Skipping scripts file:", filePath)
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

// Reads the Go file and extracts individual function bodies
func extractFunctions(filePath string, functionNames []string) (map[string]string, error) {
	if strings.Contains(filePath, "vendor/") {
		fmt.Println("Skipping vendor file:", filePath)
		return nil, nil
	}
	if strings.Contains(filePath, "scripts/") {
		fmt.Println("Skipping scripts file:", filePath)
		return nil, nil
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, filePath, content, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	functions := make(map[string]string)
	for _, funcName := range functionNames {
		ast.Inspect(node, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok && funcName == fn.Name.Name {
				start := fs.Position(fn.Pos()).Offset
				end := fs.Position(fn.End()).Offset
				functions[fn.Name.Name] = string(content[start:end])
			}
			return true
		})
	}
	if len(functions) == 0 {
		return nil, fmt.Errorf("no functions found in %s", filePath)
	}

	return functions, nil
}

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

func generateTests(client *openai.Client, functions map[string]string) string {
	var generatedTests []string

	count := 1
	for name, code := range functions {
		if count > 1 {
			break
		}
		fmt.Printf("🚀 Generating test for function: %s\n", name)
		// Construct prompt
		prompt := fmt.Sprintf(`Write a Golang unit test using "testing" package for the following function:
%s
Return only the Go unit test code inside a code block. Use a table-driven test format.Output only the code`, code)

		// Call OpenAI API
		resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
			Model: openai.GPT4Turbo,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: "You are an expert Golang developer."},
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
		})
		if err != nil {
			log.Printf("❌ Error generating test for %s: %v", name, err)
			continue
		}

		// Extract and add the test function
		generatedTests = append(generatedTests, resp.Choices[0].Message.Content)
		count++
	}
	return strings.Join(generatedTests, "\n\n")
}

// extractGoCode extracts text between ```go and ```
func extractCode(input string) string {
	var builder strings.Builder
	inBlock := false

	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "```go" {
			inBlock = true
			continue
		} else if trimmed == "```" {
			inBlock = false
			continue
		}
		if inBlock {
			builder.WriteString(line + "\n")
		}
	}

	return builder.String()
}
func main() {
	filePath := os.Args[1]
	modifiedFunctions := extractModifiedFunctions(filePath)

	fmt.Println("Modified functions:", modifiedFunctions)

	diffFunc, err := extractFunctions(filePath, modifiedFunctions)
	if err != nil {
		log.Fatalf("Error extracting test file: %v", err)
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	testCode := generateTests(client, diffFunc)

	if len(testCode) == 0 {
		return
	}
	testCode = extractCode(testCode)
	baseName := strings.TrimSuffix(filePath, ".go")
	testFile := fmt.Sprintf("%s_test.go", baseName)
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		file, err := os.Create(testFile)
		if err != nil {
			log.Fatalf("Error creating test file: %v", err)
		}
		defer file.Close() // Ensure file is closed properly
		fmt.Println("File created:", testFile)
	} else {
		fmt.Println("File already exists:", testFile)
	}
	file, err := os.OpenFile(testFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening test file: %v", err)
	}
	defer file.Close()

	// Write test code to file
	bytesWritten, err := file.WriteString(testCode)
	if err != nil {
		log.Fatalf("Error writing test file: %v", err)
	}

	// Ensure data is flushed to disk
	err = file.Sync()
	if err != nil {
		log.Fatalf("Error syncing test file: %v", err)
	}
	fmt.Printf("Test file written successfully! Bytes written: %d\n", bytesWritten)
}
