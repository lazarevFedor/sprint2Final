package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	obj "orchestrator/internal/entities"
	"orchestrator/internal/parser"
	logger2 "pkg/logger"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	UpdateExpressionStatus = "UPDATE expressions SET user_id= $1, result = $2, status = $3 WHERE id = $4"
	secretKey              = "secret"
)

// syncDBWithCache starts synchronization DB with cache
func syncDBWithCache(ctx context.Context, db *sql.DB) error {
	logger := logger2.GetLogger(ctx)
	rows, err := db.QueryContext(ctx, "SELECT id, user_id, expression FROM expressions WHERE status = $1", "In progress")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error("Error in syncDBWithCache: ", "err", err)
		return fmt.Errorf("syncDBWithCache: %w", err)
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			logger.Error("syncDBWithCache: closing rows: ", "err", err)
		}
	}(rows)
	if rows != nil {
		for rows.Next() {
			var expr string
			var userId int
			var id int
			err = rows.Scan(&id, &userId, &expr)
			if err != nil {
				logger.Error("Error in syncDBWithCache: ", "err", err.Error())
				return fmt.Errorf("syncDBWithCache: %w", err)
			}
			obj.Wg.Add(1)
			go parser.Parse(expr, id, userId)
		}
	}
	return nil
}

// startUpdatingDB updates DB with expressions in map and deletes old expressions
func startUpdatingDB(ctx context.Context, db *sql.DB) {
	ticker := time.NewTicker(15 * time.Second)
	logger := logger2.GetLogger(ctx)
	go func() {
		for range ticker.C {
			expressionsMap := obj.Expressions.GetAll()
			for key, expr := range expressionsMap {
				task, ok := expr.(obj.ClientResponse)
				if ok && (task.Status == "Done" || task.Status == "Fail") {
					_, err := db.ExecContext(ctx, UpdateExpressionStatus, task.GetUserId(), task.Result, task.Status, task.Id)
					if err != nil {
						logger.Error("Error updating expressions: ", "err", err)
						return
					}
					obj.Expressions.Delete(key)
				}
			}
		}
	}()
}

// isValidExpression checks if the expression is valid
func isValidExpression(expression string) bool {
	re := regexp.MustCompile("^[\\d+\\-*/\\s()]+$")
	return re.MatchString(expression)
}

// calculateHandler handles the /api/v1/calculate endpoint
func calculateHandler(ctx context.Context, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logger2.GetLogger(ctx)
		var clientRequest obj.ClientRequest
		var clientResponse obj.ClientResponse
		err := json.NewDecoder(r.Body).Decode(&clientRequest)
		if err != nil {
			clientResponse.Error = "Internal server error"
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("calculateHandler: could not decode request:", "err", err)
			return
		}
		if !isValidExpression(clientRequest.Expression) {
			clientResponse.Error = "Expression is not valid"
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		userId, ok := r.Context().Value("user_id").(int)
		if !ok {
			logger.Warn("calculateHandler: could not get user_id from context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		row := db.QueryRowContext(ctx, "INSERT INTO expressions(user_id, expression, status) VALUES(?, ?, ?) RETURNING id", userId, clientRequest.Expression, "In progress")
		err = row.Scan(&clientResponse.Id)
		if err != nil {
			logger.Warn("calculateHandler: could not insert expressions: ", "err", err)
			return
		}
		obj.Wg.Add(1)
		go parser.Parse(clientRequest.Expression, clientResponse.Id, userId)

		logger.Info("calculateHandler: expression was added to the queue:", "Id", clientResponse.Id)
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(clientResponse); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("calculateHandler: could not encode response:", "err", err)
			return
		}
	}
}

// expressionHandler handles the /api/v1/expressions endpoint
func expressionHandler(ctx context.Context, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logger2.GetLogger(ctx)
		userID, ok := r.Context().Value("user_id").(int)
		if !ok {
			logger.Warn("calculateHandler: could not get user_id from context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		rows, err := db.QueryContext(ctx, `
            SELECT id, status, result 
            FROM expressions 
            WHERE user_id = ?`,
			userID,
		)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Error("database query error", "error", err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				logger.Error("database close error", "error", err)
				return
			}
		}(rows)

		var expressions []obj.ClientResponse
		for rows.Next() {
			var expr obj.ClientResponse

			_ = rows.Scan(
				&expr.Id,
				&expr.Status,
				&expr.Result,
			)

			expressions = append(expressions, expr)
		}

		if err = rows.Err(); err != nil {
			logger.Error("rows iteration error", "error", err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"expressions": expressions,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("JSON encode error", "error", err)
		}
	}
}

// expressionIDHandler handles the /api/v1/expressions/{id} endpoint
func expressionIDHandler(ctx context.Context, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logger2.GetLogger(ctx)
		userID, ok := r.Context().Value("user_id").(int)
		if !ok {
			logger.Warn("calculateHandler: could not get user_id from context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		urlParts := strings.Split(r.URL.Path, "/")
		idStr := urlParts[len(urlParts)-1]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusInternalServerError)
			return
		}
		var expr obj.ClientResponse
		row := db.QueryRow(`
            SELECT result, status 
            FROM expressions 
            WHERE user_id = ? and id = ?`,
			userID, id,
		)
		err = row.Scan(&expr.Result, &expr.Status)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Error("database query error", "error", err)
			return
		}
		expr.Id = id
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(expr); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			logger.Error("expressionIDHandler: could not encode response:", "err", err)
			return
		}
	}
}

// authMiddleware checks auth status of user by jwt token
func authMiddleware(ctx context.Context) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				sendJSONError(w, "Authorization header required", http.StatusUnauthorized, ctx)
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				sendJSONError(w, "Invalid token format", http.StatusUnauthorized, ctx)
				return
			}
			tokenStr := tokenParts[1]

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secretKey), nil
			})

			if err != nil || !token.Valid {
				sendJSONError(w, "Invalid token", http.StatusUnauthorized, ctx)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				sendJSONError(w, "Invalid token claims", http.StatusUnauthorized, ctx)
				return
			}

			userID, ok := claims["user_id"].(float64)
			if !ok || userID == 0 {
				sendJSONError(w, "Invalid user ID", http.StatusUnauthorized, ctx)
				return
			}

			ctx = context.WithValue(ctx, "user_id", int(userID))
			next(w, r.WithContext(ctx))
		}
	}
}

// sendJSONError sends errors by json in authMiddleware func
func sendJSONError(w http.ResponseWriter, message string, code int, ctx context.Context) {
	logger := logger2.GetLogger(ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(map[string]string{
		"error":   http.StatusText(code),
		"message": message,
	})
	if err != nil {
		logger.Error("JSON encode error", "error", err)
		return
	}
	logger.Error("JSON encode error", "error", message)
}

// registerHandler handles the /api/v1/register endpoint
func registerHandler(ctx context.Context, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logger2.GetLogger(ctx)
		var user obj.RegisterRequest
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			logger.Error("registerHandler: could not decode request:", "err", err)
			return
		}

		_, err = db.ExecContext(ctx, "INSERT INTO users (login, password) VALUES (?, ?)", user.Login, user.Password)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			logger.Error("registerHandler: could not insert user:", "err", err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// GenerateToken generates jwt token
func GenerateToken(userID int, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}
	return tokenString, nil
}

// loginHandler handles the /api/v1/login endpoint
func loginHandler(ctx context.Context, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logger2.GetLogger(ctx)
		var user obj.LoginRequest
		var id int
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			logger.Error("loginHandler: could not decode request:", "err", err)
			return
		}
		row := db.QueryRow("SELECT id FROM users WHERE login = ? AND password = ?", user.Login, user.Password)
		err = row.Scan(&id)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			logger.Error("loginHandler: could not find user:", "err", err)
			return
		}

		token, err := GenerateToken(id, secretKey)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			logger.Error("loginHandler: could not generate token:", "err", err)
			return
		}

		resp := obj.LoginResponse{
			Token: token,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("loginHandler: could not encode response:", "err", err)
			return
		}
	}
}

// StartServer starts the server on port 8080 and listens for incoming requests
func StartServer(ctx context.Context, db *sql.DB) error {
	logger := logger2.GetLogger(ctx)

	mux := http.NewServeMux()
	//start synchronization DB with cache
	err := syncDBWithCache(ctx, db)
	if err != nil {
		return fmt.Errorf("syncDBWithCache: %w", err)
	}
	//start updating DB
	startUpdatingDB(ctx, db)
	// Handle functions for client requests
	mux.HandleFunc("/api/v1/register", registerHandler(ctx, db))
	mux.HandleFunc("/api/v1/login", loginHandler(ctx, db))
	mux.HandleFunc("/api/v1/calculate", authMiddleware(ctx)(calculateHandler(ctx, db)))
	mux.HandleFunc("/api/v1/expressions", authMiddleware(ctx)(expressionHandler(ctx, db)))
	mux.HandleFunc("/api/v1/expressions/", authMiddleware(ctx)(expressionIDHandler(ctx, db)))
	// Start the server
	logger.Info("StartServer: server started")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		logger.Error("StartServer: could not start server:", "err", err)
		return fmt.Errorf("could not start server: %v", err)
	}
	obj.Wg.Wait()
	return nil
}
