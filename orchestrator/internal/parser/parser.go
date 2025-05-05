package parser

import (
	"errors"
	"fmt"
	obj "orchestrator/internal/entities"
	"os"
	"strconv"
)

// getEnvAsInt returns the value of the environment variable as an integer
func getEnvAsInt(name string, defaultValue int) int {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// Time of operations in milliseconds
var (
	timeAdditionMs       = getEnvAsInt("TIME_ADDITION_MS", 100)
	timeSubtractionMs    = getEnvAsInt("TIME_SUBTRACTION_MS", 100)
	timeMultiplicationMs = getEnvAsInt("TIME_MULTIPLICATIONS_MS", 100)
	timeDivisionMs       = getEnvAsInt("TIME_DIVISIONS_MS", 100)
)

func returnTimeOfOperation(operation rune) int {
	switch operation {
	case '+':
		return timeAdditionMs
	case '-':
		return timeSubtractionMs
	case '*':
		return timeMultiplicationMs
	case '/':
		return timeDivisionMs
	default:
		return 100
	}
}

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

// getResult returns the result of the expression in Reverse Polish Notation
func getResult(output string, ch *chan float64, Id int) (float64, error) {
	var stack []node
	var current string
	var result float64
	var tempVariable float64
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
				if result == 0 && output[i] == '/' {
					return 0, errors.New("division by zero")
				}
				obj.Tasks.Enqueue(obj.Task{Id: Id, Arg1: tempVariable, Arg2: result, Operation: string(output[i]), OperationTime: returnTimeOfOperation(rune(output[i]))})
				result = <-*ch
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

// Parse the expression into Reverse Polish Notation and returns the result
func Parse(expression string, Id int, userId int) {
	defer obj.Wg.Done()
	var stack []node
	var output, current string
	parserChan := make(chan float64)
	t := obj.ClientResponse{
		Id:     Id,
		Status: "In progress",
	}
	t.SetUserId(userId)
	obj.Expressions.Set(strconv.Itoa(Id), t)
	fmt.Printf("Task with id(%d) and user_id(%d) has been added to the queue)", Id, userId)
	obj.ParserMutex.Lock()
	obj.ParsersTree.Insert(Id, &parserChan)
	obj.ParserMutex.Unlock()
	if expression == "" {
		obj.ParserMutex.Lock()
		_ = obj.ParsersTree.Delete(Id)
		obj.ParserMutex.Unlock()
		t.Id = Id
		t.Status = "Fail"
		t.Error = "empty expression"
		t.SetUserId(userId)
		obj.Expressions.Set(strconv.Itoa(Id), t)
		fmt.Printf("Task with id(%d) failed with error %s", Id, "empty expression")
		return
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
						obj.ParserMutex.Lock()
						_ = obj.ParsersTree.Delete(Id)
						obj.ParserMutex.Unlock()
						t.Id = Id
						t.Status = "Fail"
						t.Error = "'(' not found"
						t.SetUserId(userId)
						obj.Expressions.Set(strconv.Itoa(Id), t)
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
			obj.ParserMutex.Lock()
			_ = obj.ParsersTree.Delete(Id)
			obj.ParserMutex.Unlock()
			t.Id = Id
			t.Status = "Fail"
			t.Error = "wrong symbol"
			t.SetUserId(userId)
			obj.Expressions.Set(strconv.Itoa(Id), t)
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
	result, err := getResult(output, &parserChan, Id)
	obj.ParserMutex.Lock()
	_ = obj.ParsersTree.Delete(Id)
	obj.ParserMutex.Unlock()
	if err != nil {
		t.Id = Id
		t.Status = "Fail"
		t.Error = err.Error()
		t.SetUserId(userId)
		obj.Expressions.Set(strconv.Itoa(Id), t)
		return
	}
	t.Id = Id
	t.Status = "Done"
	t.Result = result
	t.SetUserId(userId)
	obj.Expressions.Set(strconv.Itoa(Id), t)
}
