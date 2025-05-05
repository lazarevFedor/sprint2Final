package main

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"net"
	"orchestrator/internal/grpc_server"
	"orchestrator/internal/server"
	"os"
	"pkg/api"
	"pkg/logger"
	"sync"
)

func createTables(ctx context.Context, db *sql.DB) error {
	const (
		usersTable = "CREATE TABLE IF NOT EXISTS users(id INTEGER PRIMARY KEY AUTOINCREMENT, login TEXT UNIQUE NOT NULL, password TEXT NOT NULL);"

		expressionsTable = "CREATE TABLE IF NOT EXISTS expressions(id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, expression TEXT NOT NULL, result REAL, status TEXT NOT NULL);"
	)

	if _, err := db.ExecContext(ctx, usersTable); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, expressionsTable); err != nil {
		return err
	}

	return nil
}

func main() {
	serverLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := logger.WithLogger(context.Background(), serverLogger)
	log := logger.GetLogger(ctx)

	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		log.Error("error opening sqlite3 db: ", err)
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Error("error closing sqlite3 db: ", err)
			return
		}
	}(db)

	err = db.PingContext(ctx)
	if err != nil {
		log.Error("error pinging sqlite3 db: ", err)
		return
	}

	if err = createTables(ctx, db); err != nil {
		log.Error("error creating tables: ", err)
		return
	}

	log.Info("DB created")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = server.StartServer(ctx, db)
		if err != nil {
			log.Error("error starting server: ", err)
			return
		}
		log.Info("Server started")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", ":8081")
		if err != nil {
			log.Error("error starting grpc server:", "err", err)
		}
		log.Info("starting grpc server:", "port", lis.Addr().(*net.TCPAddr).Port)
		srv := grpc_server.New()
		newServer := grpc.NewServer()
		api.RegisterOrchestratorServer(newServer, srv)
		reflection.Register(newServer)
		if err := newServer.Serve(lis); err != nil {
			log.Error("error starting grpc server:", "err", err)
		}
	}()
	wg.Wait()
}
