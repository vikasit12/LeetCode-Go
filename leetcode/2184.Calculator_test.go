package calculator

import (
	"fmt"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        float64
		operator string
		want     float64
		wantErr  bool
	}{
		{"Addition", 5, 3, "+", 8, false},
		{"Subtraction", 10, 4, "-", 6, false},
		{"Multiplication", 7, 6, "*", 42, false},
		{"Division", 8, 2, "/", 4, false},
		{"Division by zero", 5, 0, "/", 0, true},
		{"Unsupported operator", 4, 5, "%", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculate(tt.a, tt.b, tt.operator)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("calculate() = %v, want %v", got, tt.want)
			}
		})
	}
}
