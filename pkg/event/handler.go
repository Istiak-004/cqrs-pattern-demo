package event

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/istiak-004/test/go/src/github.com/learn_microservices/cqrs_pattern/pkg/models"
	"github.com/segmentio/kafka-go"
)

type EventHandler struct {
	QueryDB *sql.DB
}

func (h *EventHandler) HandleUserCreated(ctx context.Context, msg kafka.Message) error {
	var user models.User
	if err := json.Unmarshal(msg.Value, &user); err != nil {
		return err
	}

	// Update query database
	_, err := h.QueryDB.ExecContext(ctx,
		`INSERT INTO user_views (id, full_name, email, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET full_name = $2, email = $3`,
		user.ID,
		user.FirstName+" "+user.LastName,
		user.Email,
		user.CreatedAt,
	)
	if err != nil {
		return err
	}

	log.Printf("Updated user view for %s", user.ID)
	return nil
}
