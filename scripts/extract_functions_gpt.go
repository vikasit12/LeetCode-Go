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

// Reads the Go file and extracts individual function bodies
func extractFunctions(filePath string, functionNames []string) (map[string]string, error) {
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
	// if len(functions) == 0 {
	// 	return nil, fmt.Errorf("no functions found in %s", filePath)
	// }

	return functions, nil
}

// Call GPT-4 for PR analysis
// func getGPT4Analysis(codeDiff string) string {
// 	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

// 	resp, err := client.CreateChatCompletion(
// 		context.Background(),
// 		openai.ChatCompletionRequest{
// 			Model: openai.GPT4,
// 			Messages: []openai.ChatCompletionMessage{
// 				{
// 					Role:    "system",
// 					Content: "You are a Go language expert. Analyze the following code diff and identify key function changes along with their impact.",
// 				},
// 				{
// 					Role:    "user",
// 					Content: codeDiff,
// 				},
// 			},
// 		},
// 	)
// 	if err != nil {
// 		log.Fatalf("Error calling OpenAI API: %v", err)
// 	}

//		return resp.Choices[0].Message.Content
//	}
//
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
		log.Printf("abhi:---func name:%v,code : %+v", name, code)
		// Construct prompt
		prompt := fmt.Sprintf(`Write a Golang unit test using "testing" package for the following function:
%s
Return only the Go code inside a code block. Use a table-driven test format.`, code)

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
func main() {
	filePath := os.Args[1]
	modifiedFunctions := extractModifiedFunctions(filePath)

	fmt.Println("Modified functions:", modifiedFunctions)

	// diffFile := "diff.patch"
	// diffContent, err := os.ReadFile(diffFile)
	// if err != nil {
	// 	log.Fatalf("Error reading diff file: %v", err)
	// }
	diffFunc, err := extractFunctions(filePath, modifiedFunctions)
	if err != nil {
		log.Fatalf("Error extracting test file: %v", err)
	}

	// gptAnalysis := getGPT4Analysis(string(diffContent))
	// fmt.Println("GPT-4 Analysis:\n", gptAnalysis)

	// testCode := generateTestCases(diffFunc, functionCode)
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	testCode := generateTests(client, diffFunc)
	testFile := fmt.Sprintf("%s_test.go", "Checking")
	err = os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		log.Fatalf("Error writing test file: %v", err)
	}

	fmt.Println("Generated test case for:", diffFunc)
}
