package server

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	logger2 "pkg/logger"
	"strings"
	"testing"
)

// TestIsValidExpression tests the isValidExpression function
func TestIsValidExpression(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		expected   bool
	}{
		{"valid simple", "2 + 3", true},
		{"valid with parentheses", "4 * (5 - 2)", true},
		{"invalid character", "2 + a", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidExpression(tt.expression)
			assert.Equal(t, tt.expected, got, "isValidExpression(%q) = %v, want %v", tt.expression, got, tt.expected)
		})
	}
}

// TestCalculateHandler_InvalidExpression tests calculateHandler with an invalid expression
func TestCalculateHandler_InvalidExpression(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	ctx := logger2.WithLogger(context.Background(), slog.New(slog.NewJSONHandler(nil, nil)))
	token, _ := GenerateToken(1, secretKey)

	reqBody := `{"expression": "2 + a"}`
	req, _ := http.NewRequest("POST", "/api/v1/calculate", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req = req.WithContext(context.WithValue(ctx, "user_id", 1))

	rr := httptest.NewRecorder()
	handler := calculateHandler(ctx, db)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

// TestAuthMiddleware_ValidToken tests authMiddleware with a valid token
func TestAuthMiddleware_ValidToken(t *testing.T) {
	token, _ := GenerateToken(1, secretKey)
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("user_id").(int)
		assert.True(t, ok)
		assert.Equal(t, 1, userID)
		w.WriteHeader(http.StatusOK)
	})

	middleware := authMiddleware(context.Background())
	handler := middleware(dummyHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestSyncDBWithCache tests syncDBWithCache
func TestSyncDBWithCache(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "user_id", "expression"}).
		AddRow(1, 1, "2 + 3").
		AddRow(2, 1, "4 * 5")
	mock.ExpectQuery("SELECT id, user_id, expression FROM expressions WHERE status = ?").
		WithArgs("In progress").
		WillReturnRows(rows)

	ctx := logger2.WithLogger(context.Background(), slog.New(slog.NewJSONHandler(nil, nil)))
	err = syncDBWithCache(ctx, db)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
