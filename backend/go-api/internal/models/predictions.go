package models

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type PredictionHistory struct {
	ID             uuid.UUID       `json:"id"`
	UserID         uuid.UUID       `json:"user_id"`
	ModelID        *uuid.UUID      `json:"model_id,omitempty"`
	PredictionType string          `json:"prediction_type"`
	InputData      json.RawMessage `json:"input_data"`
	Prediction     json.RawMessage `json:"prediction"`
	Confidence     *float64        `json:"confidence,omitempty"`
	CreatedAt      string          `json:"created_at"`
}

func (r *Repository) GetPredictionHistory(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	offset int,
) ([]PredictionHistory, error) {

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	rows, err := r.DB.Query(ctx, `
		SELECT
			id,
			user_id,
			model_id,
			prediction_type,
			input_data,
			prediction,
			confidence,
			created_at
		FROM prediction_history
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]PredictionHistory, 0)

	for rows.Next() {
		var item PredictionHistory

		var createdAt time.Time
		err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.ModelID,
			&item.PredictionType,
			&item.InputData,
			&item.Prediction,
			&item.Confidence,
			&createdAt,
		)

		if err != nil {
			return nil, err
		}

item.CreatedAt = createdAt.Format(time.RFC3339)
		history = append(history, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}
