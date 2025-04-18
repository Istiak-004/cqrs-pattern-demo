package query

import (
	"context"
	"database/sql"

	"github.com/istiak-004/test/go/src/github.com/learn_microservices/cqrs_pattern/pkg/models"
)

type QueryHandler struct {
	QueryDB *sql.DB
}

func (h *QueryHandler) GetUser(ctx context.Context, id string) (*models.UserView, error) {
	var user models.UserView
	err := h.QueryDB.QueryRowContext(ctx,
		`SELECT id, full_name, email, created_at 
		FROM user_views 
		WHERE id = $1`, id).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (h *QueryHandler) ListUsers(ctx context.Context) ([]models.UserView, error) {
	rows, err := h.QueryDB.QueryContext(ctx,
		`SELECT id, full_name, email, created_at 
		FROM user_views 
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.UserView
	for rows.Next() {
		var user models.UserView
		if err := rows.Scan(
			&user.ID,
			&user.FullName,
			&user.Email,
			&user.CreatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
