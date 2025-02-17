package calc

import "testing"

func TestPriority(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{input: "+", want: 1},
		{input: "-", want: 1},
		{input: "*", want: 2},
		{input: "/", want: 2},
		{input: "(", want: 0},
		{input: ")", want: 0},
		{input: "1", want: -1},
		{input: "a", want: -2},
	}
	for _, tt := range tests {
		if got := priority(tt.input); got != tt.want {
			t.Errorf("priority(%s): expected %d, got %d", tt.input, tt.want, got)
		}
	}
}

func TestOperation(t *testing.T) {
	tests := []struct {
		a, b    float64
		op      string
		want    float64
		wantErr bool
	}{
		{a: 1, b: 2, op: "+", want: 3, wantErr: false},
		{a: 1, b: 2, op: "-", want: 1, wantErr: false},
		{a: 2, b: 2, op: "*", want: 4, wantErr: false},
		{a: 2, b: 2, op: "/", want: 1, wantErr: false},
		{a: 0, b: 2, op: "/", want: 0, wantErr: true},
		{a: 1, b: 2, op: "%", want: 0, wantErr: true},
	}
	for _, tt := range tests {
		got, err := operation(tt.a, tt.b, tt.op)
		if (err != nil) != tt.wantErr {
			t.Errorf("operation(%f, %f, %s): expected error %v, got %v", tt.a, tt.b, tt.op, tt.wantErr, err)
		}
		if got != tt.want {
			t.Errorf("operation(%f, %f, %s): expected %f, got %f", tt.a, tt.b, tt.op, tt.want, got)
		}
	}
}

func TestGetResult(t *testing.T) {
	tests := []struct {
		input   string
		want    float64
		wantErr bool
	}{
		{input: "2 2 +", want: 4, wantErr: false},
		{input: "2 2 -", want: 0, wantErr: false},
		{input: "2 2 *", want: 4, wantErr: false},
		{input: "2 2 /", want: 1, wantErr: false},
		{input: "2 0 /", want: 0, wantErr: true},
		{input: "2 a +", want: 0, wantErr: true},
		{input: "2 2 + +", want: 0, wantErr: true},
		{input: "2 2 + 3", want: 0, wantErr: true},
	}
	for _, tt := range tests {
		got, err := getResult(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("getResult(%s): expected error %v, got %v", tt.input, tt.wantErr, err)
		}
		if got != tt.want {
			t.Errorf("getResult(%s): expected %f, got %f", tt.input, tt.want, got)
		}
	}
}

func TestCalcErrors(t *testing.T) {
	tests := []struct {
		expression string
		wantErr    bool
	}{
		{expression: "2+a", wantErr: true},
		{expression: "2+(2", wantErr: true},
		{expression: "2+2)", wantErr: true},
		{expression: "2+2+", wantErr: true},
		{expression: "2+2+2", wantErr: false},
		{expression: "", wantErr: true},
		{expression: "2", wantErr: false},
		{expression: "2/0", wantErr: true},
		{expression: "2++", wantErr: true},
		{expression: "2 2", wantErr: true},
		{expression: "2 2 +", wantErr: true},
		{expression: "2 2 + 2", wantErr: true},
		{expression: "(2)", wantErr: false},
		{expression: ")", wantErr: true},
	}
	for i, tt := range tests {
		_, err := Calc(tt.expression)
		if (err != nil) != tt.wantErr {
			t.Errorf("Calc(%s): expected error %v, got %v, test num = %d", tt.expression, tt.wantErr, err, i+1)
		}
	}
}
