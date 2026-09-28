package models

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

type CreateModelParams struct {
	ID         uuid.UUID
	DatasetID  *uuid.UUID
	Name       string
	ModelType  string
	TaskType   string
	Metrics    map[string]interface{}
	Parameters map[string]interface{}
}

type Model struct {
	ID         uuid.UUID              `json:"id"`
	DatasetID  *uuid.UUID             `json:"dataset_id,omitempty"`
	Name       string                 `json:"name"`
	ModelType  string                 `json:"model_type"`
	TaskType   string                 `json:"task_type"`
	Metrics    map[string]interface{} `json:"metrics"`
	Parameters map[string]interface{} `json:"parameters"`
	CreatedAt  time.Time              `json:"created_at"`
}

func (r *Repository) Create(
	ctx context.Context,
	params CreateModelParams,
) (uuid.UUID, error) {
	metrics, err := json.Marshal(params.Metrics)
	if err != nil {
		return uuid.Nil, err
	}

	parameters, err := json.Marshal(params.Parameters)
	if err != nil {
		return uuid.Nil, err
	}

	var id uuid.UUID

	err = r.DB.QueryRow(
		ctx,
		`
		INSERT INTO models (
			id,
			dataset_id,
			name,
			model_type,
			task_type,
			metrics,
			parameters
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
		`,
		params.ID,
		params.DatasetID,
		params.Name,
		params.ModelType,
		params.TaskType,
		metrics,
		parameters,
	).Scan(&id)

	return id, err
}

func (r *Repository) Exists(
	ctx context.Context,
	id uuid.UUID,
) (bool, error) {
	var exists bool

	err := r.DB.QueryRow(
		ctx,
		`
		SELECT EXISTS(
			SELECT 1
			FROM models
			WHERE id = $1
		)
		`,
		id,
	).Scan(&exists)

	return exists, err
}

func (r *Repository) List(
	ctx context.Context,
) ([]Model, error) {
	rows, err := r.DB.Query(
		ctx,
		`
		SELECT
			id,
			dataset_id,
			name,
			model_type,
			task_type,
			metrics,
			parameters,
			created_at
		FROM models
		ORDER BY created_at DESC
		`,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Model, 0)

	for rows.Next() {
		var model Model
		var metricsJSON []byte
		var parametersJSON []byte

		err := rows.Scan(
			&model.ID,
			&model.DatasetID,
			&model.Name,
			&model.ModelType,
			&model.TaskType,
			&metricsJSON,
			&parametersJSON,
			&model.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		if len(metricsJSON) > 0 {
			if err := json.Unmarshal(metricsJSON, &model.Metrics); err != nil {
				return nil, err
			}
		}

		if len(parametersJSON) > 0 {
			if err := json.Unmarshal(parametersJSON, &model.Parameters); err != nil {
				return nil, err
			}
		}

		result = append(result, model)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*Model, error) {
	var model Model
	var metricsJSON []byte
	var parametersJSON []byte

	err := r.DB.QueryRow(
		ctx,
		`
		SELECT
			id,
			dataset_id,
			name,
			model_type,
			task_type,
			metrics,
			parameters,
			created_at
		FROM models
		WHERE id = $1
		`,
		id,
	).Scan(
		&model.ID,
		&model.DatasetID,
		&model.Name,
		&model.ModelType,
		&model.TaskType,
		&metricsJSON,
		&parametersJSON,
		&model.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if len(metricsJSON) > 0 {
		if err := json.Unmarshal(metricsJSON, &model.Metrics); err != nil {
			return nil, err
		}
	}

	if len(parametersJSON) > 0 {
		if err := json.Unmarshal(parametersJSON, &model.Parameters); err != nil {
			return nil, err
		}
	}

	return &model, nil
}

type SavePredictionHistoryParams struct {
	UserID         uuid.UUID
	ModelID        uuid.UUID
	PredictionType string
	InputData      []byte
	Prediction     []byte
	Confidence     *float64
}

func (r *Repository) SavePredictionHistory(
	ctx context.Context,
	params SavePredictionHistoryParams,
) (uuid.UUID, error) {
	var id uuid.UUID

	err := r.DB.QueryRow(
		ctx,
		`
		INSERT INTO prediction_history (
			user_id,
			model_id,
			prediction_type,
			input_data,
			prediction,
			confidence
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
		`,
		params.UserID,
		params.ModelID,
		params.PredictionType,
		params.InputData,
		params.Prediction,
		params.Confidence,
	).Scan(&id)

	return id, err
}
