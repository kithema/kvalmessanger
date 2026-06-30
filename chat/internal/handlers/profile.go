package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"2say/internal/middleware"
	"2say/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func (h Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user models.User
	err := h.DB.QueryRow(
		"SELECT id, alias, name, photo_url FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Alias, &user.Name, &user.PhotoURL)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)

	if req.Name != "" {
		_, err := h.DB.Exec(
			"UPDATE users SET name = $1 WHERE id = $2",
			req.Name, userID,
		)
		if err != nil {
			http.Error(w, "Cannot update profile", http.StatusInternalServerError)
			return
		}
	}

	if req.Password != "" {
		if len(req.Password) < 6 {
			http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
			return
		}
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		_, err = h.DB.Exec(
			"UPDATE users SET password_hash = $1 WHERE id = $2",
			passwordHash, userID,
		)
		if err != nil {
			http.Error(w, "Cannot update password", http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}