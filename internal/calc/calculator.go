package calc

import (
	"errors"
	"strconv"
)

// node is a struct that contains data and priority
type node struct {
	Data     string
	Priority int
}

// priority returns the priority of the operator
func priority(c string) int {
	switch c {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	case "(", ")":
		return 0
	default:
		if c >= "0" && c <= "9" {
			return -1
		} else {
			return -2
		}
	}
}

// operation performs the operation of two numbers
func operation(a, b float64, s string) (float64, error) {
	switch s {
	case "+":
		return b + a, nil
	case "-":
		return b - a, nil
	case "*":
		return b * a, nil
	case "/":
		if a == 0 {
			return 0, errors.New("division by zero")
		}
		return b / a, nil
	default:
		return 0, errors.New("wrong operator")
	}
}

// getResult returns the result of the expression in Reverse Polish Notation
func getResult(output string) (float64, error) {
	var stack []node
	var current string
	var result float64
	var tempVariable float64
	var err error
	for i := 0; i < len(output); i++ {
		if output[i] != ' ' {
			switch priority(string(output[i])) {
			case -1:
				current += string(output[i])
			case 0:
				return 0, errors.New("error '(' or ')' in output string")
			case 1, 2:
				if len(stack) < 2 {
					return 0, errors.New("out of operands")
				}
				result, _ = strconv.ParseFloat(stack[len(stack)-1].Data, 64)
				stack = append(stack[:len(stack)-1])
				tempVariable, _ = strconv.ParseFloat(stack[len(stack)-1].Data, 64)
				stack = append(stack[:len(stack)-1])
				result, err = operation(result, tempVariable, string(output[i]))
				if err != nil {
					return 0, err
				}
				stack = append(stack, node{Data: strconv.FormatFloat(result, 'f', 2, 64)})
				current = ""
			default:
				return 0, errors.New("wrong symbol")
			}
		} else {
			if current != "" {
				stack = append(stack, node{Data: current})
				current = ""
			}
		}
	}
	if current != "" {
		stack = append(stack, node{Data: current})
		current = ""
	}
	if len(stack) > 1 {
		return 0, errors.New("stack contains elements")
	}
	result, _ = strconv.ParseFloat(stack[0].Data, 64)
	return result, nil
}

// Calc parse the expression into Reverse Polish Notation and returns the result
func Calc(expression string) (float64, error) {
	var stack []node
	var output, current string
	if expression == "" {
		return 0, errors.New("empty expression")
	}
	for i := 0; i < len(expression); i++ {
		switch priority(string(expression[i])) {
		case -1:
			current += string(expression[i])
		case 0:
			if current != "" {
				output += current + " "
				current = ""
			}
			if expression[i] == '(' {
				stack = append(stack, node{Data: "(", Priority: 0})
			} else {
				for {
					if len(stack) == 0 {
						return 0, errors.New("'(' not found")
					}
					if stack[len(stack)-1].Data == "(" {
						break
					}
					output += stack[len(stack)-1].Data + " "
					stack = append(stack[:len(stack)-1])
				}
				stack = append(stack[:len(stack)-1])
			}
		case 1, 2:
			if current != "" {
				output += current + " "
				current = ""
			}
			if len(stack) == 0 || stack[len(stack)-1].Priority < priority(string(expression[i])) {
				stack = append(stack, node{Data: string(expression[i]), Priority: priority(string(expression[i]))})
			} else {
				for len(stack) != 0 && stack[len(stack)-1].Priority >= priority(string(expression[i])) {
					output += stack[len(stack)-1].Data + " "
					stack = append(stack[:len(stack)-1])
				}
				stack = append(stack, node{Data: string(expression[i]), Priority: priority(string(expression[i]))})
			}
		default:
			return 0, errors.New("wrong symbol")
		}
	}
	if current != "" {
		output += current + " "
		current = ""
	}
	for len(stack) != 0 {
		output += stack[len(stack)-1].Data + " "
		stack = append(stack[:len(stack)-1])
	}
	result, err := getResult(output)
	if err != nil {
		return 0, err
	}
	return result, nil
}
