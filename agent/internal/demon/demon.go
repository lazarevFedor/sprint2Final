package demon

import (
	"errors"
	"time"
)

// CalculateExpression calculates the expression
func CalculateExpression(a, b float64, operation string, operationTime int) (float64, error) {
	switch operation {
	case "+":
		time.Sleep(time.Duration(operationTime) * time.Second)
		return a + b, nil
	case "-":
		time.Sleep(time.Duration(operationTime) * time.Second)
		return a - b, nil
	case "*":
		time.Sleep(time.Duration(operationTime) * time.Second)
		return a * b, nil
	case "/":
		time.Sleep(time.Duration(operationTime) * time.Second)
		return a / b, nil
	default:
		return 0, errors.New("wrong operator")
	}
}
