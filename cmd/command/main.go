package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/istiak-004/test/go/src/github.com/learn_microservices/cqrs_pattern/pkg/command"
	"github.com/istiak-004/test/go/src/github.com/learn_microservices/cqrs_pattern/pkg/event"
	"github.com/istiak-004/test/go/src/github.com/learn_microservices/cqrs_pattern/pkg/models"
)

func main() {
	// Connect to command database
	commandDB, err := sql.Open("postgres", "postgres://cqrs:cqrs123@localhost:5432/cqrs_command?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer commandDB.Close()

	// Initialize event bus
	eventBus, err := event.NewKafkaEventBus([]string{"localhost:9092"})
	if err != nil {
		log.Fatal(err)
	}
	defer eventBus.Close()
	// Initialize command handler
	handler := command.CommandHandler{
		CommandDB: commandDB,
		EventBus:  eventBus,
	}
	// Create tables
	_, err = commandDB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`)
	if err != nil {
		log.Fatal(err)
	}

	// HTTP server
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var cmd models.CreateUserCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id, err := handler.CreateUser(r.Context(), cmd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": id})
	})
	// Start server
	server := &http.Server{Addr: ":8080", Handler: nil}
	go func() {
		log.Println("Command service started on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Command service stopped")
}
