package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"2say/internal/jwt"
	"2say/internal/models"


	"golang.org/x/crypto/bcrypt"
)

func (h Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var regreq models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&regreq); err != nil {
		http.Error(w, "error data format", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	regreq.Alias = strings.TrimSpace(regreq.Alias)
	regreq.Name = strings.TrimSpace(regreq.Name)
	if len(regreq.Alias) < 4 || len(regreq.Password) < 6 {
		http.Error(w, "Password may be >= 6 and Alias may be >= 4", http.StatusBadRequest)
		return
	}
	if regreq.Name == "" {
		regreq.Name = regreq.Alias
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(regreq.Password), 12)
	if err != nil {
		log.Printf("Cannot hash password: %s\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	var exists bool
	err = h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE alias=$1)", regreq.Alias).Scan(&exists)
	if err != nil {
		log.Printf("%s\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Alias already used", http.StatusConflict)
		return
	}
	_, err = h.DB.Exec("INSERT INTO users (alias, name, password_hash) VALUES ($1, $2, $3)", regreq.Alias, regreq.Name, passwordHash)
	if err != nil {
		log.Printf("Error exec db: %s", err)
		http.Error(w, "Cannot create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Register successful"})
}

func (h Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var logreq models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&logreq); err != nil {
		http.Error(w, "error data format", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	logreq.Alias = strings.TrimSpace(logreq.Alias)
	logreq.Password = strings.TrimSpace(logreq.Password)

	var userID int
	var passwordHash string
	var name string
	err := h.DB.QueryRow("SELECT id, password_hash, name FROM users WHERE alias=$1", logreq.Alias).Scan(&userID, &passwordHash, &name)
	if err != nil {
		log.Printf("Queryrow error%s\n", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(logreq.Password)); err != nil {
		log.Printf("Password is incorrect: %s", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := jwt.GenerateToken(userID)
	if err != nil {
		http.Error(w, "Cannot generate token", http.StatusInternalServerError)
		return
	}

	response := models.LoginResponse{
		Token: token,
	}
	response.User.ID = userID
	response.User.Alias = logreq.Alias
	response.User.Name = name

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}