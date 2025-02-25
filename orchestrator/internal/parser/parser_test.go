package parser

import "testing"

func TestReturnTimeOfOperation(t *testing.T) {
	tests := []struct {
		operation rune
		want      int
	}{
		{'+', timeAdditionMs},
		{'-', timeSubtractionMs},
		{'*', timeMultiplicationMs},
		{'/', timeDivisionMs},
	}

	for _, tt := range tests {
		t.Run(string(tt.operation), func(t *testing.T) {
			got := returnTimeOfOperation(tt.operation)
			if got != tt.want {
				t.Errorf("returnTimeOfOperation(%c) = %d; want %d", tt.operation, got, tt.want)
			}
		})
	}
}

func TestPriority(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"+", 1},
		{"-", 1},
		{"*", 2},
		{"/", 2},
		{"(", 0},
		{")", 0},
		{"1", -1},
		{"a", -2},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := priority(tt.input)
			if got != tt.want {
				t.Errorf("priority(%s) = %d; want %d", tt.input, got, tt.want)
			}
		})
	}
}

//TODO: add tests for Parse and getResult
