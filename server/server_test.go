package server

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCalculateHandler(t *testing.T) {
	tests := []struct {
		expression string
		want       string
		wantErr    bool
	}{
		{expression: `{"expression": "2+2"}`, want: `{"result":"4.00"}`},
		{expression: `{"expression": "2+A"}`, want: `{"error":"Expression is not valid"}`},
		{expression: `{"expression": "0/0"}`, want: `{"error":"Internal server error"}`},
	}
	for _, tt := range tests {
		req := httptest.NewRequest("POST", "/api/v1/calculate", nil)
		req.Body = io.NopCloser(strings.NewReader(tt.expression))
		w := httptest.NewRecorder()
		calculateHandler(w, req)
		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		bodyStr = strings.Replace(bodyStr, "\n", "", -1)
		if bodyStr != tt.want {
			t.Errorf("calculateHandler() failed, expected %v, got %v", tt.want, bodyStr)
		}
	}
}
