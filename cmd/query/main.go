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

	"github.com/istiak-004/test/go/src/github.com/learn_microservices/cqrs_pattern/pkg/event"
	"github.com/istiak-004/test/go/src/github.com/learn_microservices/cqrs_pattern/pkg/query"
	_ "github.com/lib/pq"
)

func main() {
	// Connect to query database
	queryDB, err := sql.Open("postgres", "postgres://cqrs:cqrs123@localhost:5433/cqrs_query?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer queryDB.Close()

	// Initialize query handler
	queryHandler := query.QueryHandler{
		QueryDB: queryDB,
	}

	// Initialize event handler
	eventHandler := event.EventHandler{
		QueryDB: queryDB,
	}

	// Create tables
	_, err = queryDB.Exec(`
		CREATE TABLE IF NOT EXISTS user_views (
			id TEXT PRIMARY KEY,
			full_name TEXT NOT NULL,
			email TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		)`)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Kafka consumer
	reader := event.NewKafkaReader([]string{"localhost:9092"}, "user-events", "query-service-group")
	defer reader.Close()

	// Start event consumer
	go func() {
		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}

			switch string(msg.Key) {
			case "user-created":
				if err := eventHandler.HandleUserCreated(context.Background(), msg); err != nil {
					log.Printf("Error handling user created event: %v", err)
				}
			}
		}
	}()

	// HTTP server
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			users, err := queryHandler.ListUsers(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(users)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.URL.Path[len("/users/"):]
		user, err := queryHandler.GetUser(r.Context(), id)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})

	// Start server
	server := &http.Server{Addr: ":8081", Handler: nil}
	go func() {
		log.Println("Query service started on :8081")
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
	log.Println("Query service stopped")
}
