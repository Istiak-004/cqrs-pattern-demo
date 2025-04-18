package command

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/istiak-004/test/go/src/github.com/learn_microservices/cqrs_pattern/pkg/event"
	"github.com/istiak-004/test/go/src/github.com/learn_microservices/cqrs_pattern/pkg/models"
)

type CommandHandler struct {
	CommandDB *sql.DB
	EventBus  *event.KafkaEventBus
}



func (h *CommandHandler) CreateUser(ctx context.Context, cmd models.CreateUserCommand) (string, error) {
	// Generate ID
	id := uuid.New().String()
	
	// Insert into command database
	_, err := h.CommandDB.ExecContext(ctx,
		`INSERT INTO users (id, first_name, last_name, email, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		id, cmd.FirstName, cmd.LastName, cmd.Email, time.Now(), time.Now())
	if err != nil {
		return "", err
	}

	// Publish user created event
	user := models.User{
		ID:        id,
		FirstName: cmd.FirstName,
		LastName:  cmd.LastName,
		Email:     cmd.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	eventData, err := json.Marshal(user)
	if err != nil {
		log.Printf("Failed to marshal user created event: %v", err)
		return id, nil
	}

	err = h.EventBus.Publish(ctx, "user-created", id, eventData)
	if err != nil {
		log.Printf("Failed to publish user created event: %v", err)
	}

	return id, nil
}