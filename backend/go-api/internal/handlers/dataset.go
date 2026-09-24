package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/OmaleGrace/Grace-Predict/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DatasetHandler struct {
	DB           *pgxpool.Pool
	MLServiceURL string
	UploadDir    string
}

func NewDatasetHandler(
	db *pgxpool.Pool,
	mlServiceURL string,
	uploadDir string,
) *DatasetHandler {
	return &DatasetHandler{
		DB:           db,
		MLServiceURL: mlServiceURL,
		UploadDir:    uploadDir,
	}
}

func (h *DatasetHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "user identity not found", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" {
		http.Error(w, "dataset name is required", http.StatusBadRequest)
		return
	}

	var datasetID string

	err := h.DB.QueryRow(
		r.Context(),
		`
		INSERT INTO datasets (user_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id
		`,
		userID,
		req.Name,
		req.Description,
	).Scan(&datasetID)

	if err != nil {
		http.Error(w, "failed to create dataset", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      datasetID,
		"name":    req.Name,
		"message": "dataset created successfully",
	})
}

func (h *DatasetHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "user identity not found", http.StatusUnauthorized)
		return
	}

	rows, err := h.DB.Query(
		r.Context(),
		`
		SELECT id, name, description, file_path, row_count, column_count, created_at
		FROM datasets
		WHERE user_id = $1
		ORDER BY created_at DESC
		`,
		userID,
	)
	if err != nil {
		http.Error(w, "failed to fetch datasets", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	datasets := make([]map[string]interface{}, 0)

	for rows.Next() {
		var (
			id          string
			name        string
			description *string
			filePath    *string
			rowCount    *int
			columnCount *int
			createdAt   interface{}
		)

		if err := rows.Scan(
			&id,
			&name,
			&description,
			&filePath,
			&rowCount,
			&columnCount,
			&createdAt,
		); err != nil {
			http.Error(w, "failed to read datasets", http.StatusInternalServerError)
			return
		}

		datasets = append(datasets, map[string]interface{}{
			"id":           id,
			"name":         name,
			"description":  description,
			"file_path":    filePath,
			"row_count":    rowCount,
			"column_count": columnCount,
			"created_at":   createdAt,
		})
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "failed to read datasets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(datasets)
}

func (h *DatasetHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "user identity not found", http.StatusUnauthorized)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "CSV file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		http.Error(w, "only CSV files are supported", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))

	if name == "" {
		name = strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	}

	if err := os.MkdirAll(h.UploadDir, 0755); err != nil {
		http.Error(w, "failed to create upload directory", http.StatusInternalServerError)
		return
	}

	safeFilename := filepath.Base(header.Filename)
	filePath := filepath.Join(h.UploadDir, safeFilename)

	diskFile, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	defer diskFile.Close()

	if _, err := io.Copy(diskFile, file); err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	// Re-open the saved file to send to the Python service.
	savedFile, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "failed to process file", http.StatusInternalServerError)
		return
	}
	defer savedFile.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", safeFilename)
	if err != nil {
		http.Error(w, "failed to prepare ML request", http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(part, savedFile); err != nil {
		http.Error(w, "failed to prepare ML request", http.StatusInternalServerError)
		return
	}

	if err := writer.Close(); err != nil {
		http.Error(w, "failed to prepare ML request", http.StatusInternalServerError)
		return
	}

	mlResponse, err := http.Post(
		h.MLServiceURL+"/datasets/inspect",
		writer.FormDataContentType(),
		&body,
	)
	if err != nil {
		http.Error(w, "ML service unavailable", http.StatusBadGateway)
		return
	}
	defer mlResponse.Body.Close()

	if mlResponse.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(mlResponse.Body)
		http.Error(
			w,
			fmt.Sprintf("ML service error: %s", string(message)),
			http.StatusBadGateway,
		)
		return
	}

	var inspection struct {
		Filename      string `json:"filename"`
		Rows          int    `json:"rows"`
		Columns       int    `json:"columns"`
		ColumnDetails []struct {
			Name    string `json:"name"`
			Dtype   string `json:"dtype"`
			Missing int    `json:"missing"`
			Unique  int    `json:"unique"`
		} `json:"column_details"`
	}

	if err := json.NewDecoder(mlResponse.Body).Decode(&inspection); err != nil {
		http.Error(w, "invalid response from ML service", http.StatusBadGateway)
		return
	}

	var datasetID string

	err = h.DB.QueryRow(
		r.Context(),
		`
		INSERT INTO datasets (
			user_id,
			name,
			description,
			file_path,
			row_count,
			column_count
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
		`,
		userID,
		name,
		description,
		filePath,
		inspection.Rows,
		inspection.Columns,
	).Scan(&datasetID)

	if err != nil {
		http.Error(w, "failed to save dataset", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":             datasetID,
		"name":           name,
		"filename":       inspection.Filename,
		"rows":           inspection.Rows,
		"columns":        inspection.Columns,
		"column_details": inspection.ColumnDetails,
	})
}