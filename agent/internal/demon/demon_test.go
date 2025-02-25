package demon

import (
	"testing"
	"time"
)

func TestCalculateExpression(t *testing.T) {
	tests := []struct {
		a             float64
		b             float64
		operation     string
		operationTime int
		want          float64
		wantErr       bool
	}{
		{1, 2, "+", 100, 3, false},
		{5, 3, "-", 100, 2, false},
		{2, 3, "*", 100, 6, false},
		{6, 2, "/", 100, 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			start := time.Now()
			got, err := CalculateExpression(tt.a, tt.b, tt.operation, tt.operationTime)
			duration := time.Since(start)

			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CalculateExpression() = %v, want %v", got, tt.want)
			}
			if duration < time.Duration(tt.operationTime)*time.Millisecond {
				t.Errorf("CalculateExpression() duration = %v, want at least %v", duration, time.Duration(tt.operationTime)*time.Millisecond)
			}
		})
	}
}
