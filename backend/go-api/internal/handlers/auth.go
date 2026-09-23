package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/OmaleGrace/Grace-Predict/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	DB *pgxpool.Pool
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)

	if !ok {
		http.Error(w, "user identity not found", http.StatusUnauthorized)
		return
	}

	var email string
	var name *string

	err := h.DB.QueryRow(
		r.Context(),
		`
		SELECT email, name
		FROM users
		WHERE id = $1
		`,
		userID,
	).Scan(&email, &name)

	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    userID,
		"email": email,
		"name":  name,
	})
}

func NewAuthHandler(db *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{
		DB: db,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	var userID string
	var passwordHash string
	var name *string

	err := h.DB.QueryRow(
		r.Context(),
		`
		SELECT id, password_hash, name
		FROM users
		WHERE email = $1
		`,
		req.Email,
	).Scan(&userID, &passwordHash, &name)

	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	if err := auth.CheckPassword(req.Password, passwordHash); err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(userID)
	if err != nil {
		http.Error(w, "failed to create authentication token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "login successful",
		"user_id": userID,
		"name":    name,
		"token":   token,
	})
}
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to process password", http.StatusInternalServerError)
		return
	}

	var userID string

	err = h.DB.QueryRow(
		r.Context(),
		`
		INSERT INTO users (email, password_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id
		`,
		req.Email,
		passwordHash,
		req.Name,
	).Scan(&userID)

	if err != nil {
		http.Error(w, "email may already be registered", http.StatusConflict)
		return
	}

	token, err := auth.GenerateToken(userID)
	if err != nil {
		http.Error(w, "failed to create authentication token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "user registered successfully",
		"user_id": userID,
		"token":   token,
	})
}
