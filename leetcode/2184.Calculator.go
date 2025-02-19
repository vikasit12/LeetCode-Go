package main

import (
	"fmt"
)

// Calculator function to perform basic operations
func calculate(a float64, b float64, operator string) (float64, error) {
	switch operator {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, fmt.Errorf("error: division by zero")
		}
		return a / b, nil
	default:
		return 0, fmt.Errorf("error: unsupported operator")
	}
}

func main() {
	var a, b float64
	var operator string

	fmt.Print("Enter first number: ")
	fmt.Scanln(&a)

	fmt.Print("Enter an operator (+, -, *, /): ")
	fmt.Scanln(&operator)

	fmt.Print("Enter second number: ")
	fmt.Scanln(&b)

	result, err := calculate(a, b, operator)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Result: %.2f\n", result)
	}
}
