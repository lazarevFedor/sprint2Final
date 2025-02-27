package parser

import (
	"errors"
	"testing"
)

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

func TestGetResult(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		ch       chan float64
		Id       int
		expected float64
		err      error
	}{
		{
			name:     "simple addition",
			output:   "2 3 +",
			ch:       make(chan float64, 1),
			Id:       1,
			expected: 5,
			err:      nil,
		},
		{
			name:     "division by zero",
			output:   "2 0 /",
			ch:       make(chan float64, 1),
			Id:       2,
			expected: 0,
			err:      errors.New("division by zero"),
		},
		{
			name:     "out of operands",
			output:   "2 +",
			ch:       make(chan float64, 1),
			Id:       3,
			expected: 0,
			err:      errors.New("out of operands"),
		},
		{
			name:     "wrong symbol",
			output:   "2 3 &",
			ch:       make(chan float64, 1),
			Id:       4,
			expected: 0,
			err:      errors.New("wrong symbol"),
		},
		{
			name:     "stack contains elements",
			output:   "2 3 + 4",
			ch:       make(chan float64, 1),
			Id:       5,
			expected: 0,
			err:      errors.New("stack contains elements"),
		},
		{
			name:     "error '(' or ')' in output string",
			output:   "2 3 4 (",
			ch:       make(chan float64, 1),
			Id:       5,
			expected: 0,
			err:      errors.New("error '(' or ')' in output string"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil || tt.name == "stack contains elements" {
				tt.ch <- tt.expected
			}

			result, err := getResult(tt.output, &tt.ch, tt.Id)

			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("expected error: %v, got: %v", tt.err, err)
			}

			if result != tt.expected {
				t.Errorf("expected result: %v, got: %v", tt.expected, result)
			}
		})
	}
}
