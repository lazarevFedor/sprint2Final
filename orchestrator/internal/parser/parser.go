package parser

import (
	"errors"
	obj "orchestrator/internal/entities"
	"strconv"
	"time"
)

//TODO: fix errors returning in Parse function and getResult function

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
				obj.Tasks.Enqueue(obj.Task{Id: Id, Arg1: tempVariable, Arg2: result, Operation: string(output[i])})
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
func Parse(expression string, Id int) {
	defer obj.Wg.Done()
	var stack []node
	var output, current string
	parserChan := make(chan float64)
	obj.Expressions.Set(strconv.Itoa(Id), obj.ClientResponse{Id: Id, Status: "In progress"})
	obj.ParserMutex.Lock()
	obj.ParsersTree.Insert(Id, &parserChan)
	obj.ParserMutex.Unlock()
	if expression == "" {
		obj.ParserMutex.Lock()
		obj.ParsersTree.Delete(Id)
		obj.ParserMutex.Unlock()
		obj.Expressions.Set(strconv.Itoa(Id), obj.ClientResponse{Id: Id, Status: "Fail", Error: "empty expression"})
		task := obj.Expressions.Get(strconv.Itoa(Id)).(obj.ClientResponse)
		task.SetTimestamp(time.Now())
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
						obj.ParsersTree.Delete(Id)
						obj.ParserMutex.Unlock()
						obj.Expressions.Set(strconv.Itoa(Id), obj.ClientResponse{Id: Id, Status: "Fail", Error: "'(' not found"})
						task := obj.Expressions.Get(strconv.Itoa(Id)).(obj.ClientResponse)
						task.SetTimestamp(time.Now())
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
			obj.ParsersTree.Delete(Id)
			obj.ParserMutex.Unlock()
			obj.Expressions.Set(strconv.Itoa(Id), obj.ClientResponse{Id: Id, Status: "Fail", Error: "wrong symbol"})
			task := obj.Expressions.Get(strconv.Itoa(Id)).(obj.ClientResponse)
			task.SetTimestamp(time.Now())
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
	obj.ParsersTree.Delete(Id)
	obj.ParserMutex.Unlock()
	if err != nil {
		obj.Expressions.Set(strconv.Itoa(Id), obj.ClientResponse{Id: Id, Status: "Fail", Error: err.Error()})
		task := obj.Expressions.Get(strconv.Itoa(Id)).(obj.ClientResponse)
		task.SetTimestamp(time.Now())
		return
	}
	obj.Expressions.Set(strconv.Itoa(Id), obj.ClientResponse{Id: Id, Status: "Done", Result: result})
	task := obj.Expressions.Get(strconv.Itoa(Id)).(obj.ClientResponse)
	task.SetTimestamp(time.Now())
}
