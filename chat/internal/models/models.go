package models

import "github.com/golang-jwt/jwt/v5"

type User struct {
	ID       int    `json:"id"`
	Alias    string `json:"alias"`
	Name     string `json:"name"`
	Password string `json:"password"`
	PhotoURL string `json:"photo_url"`
}

type RegisterRequest struct {
	Alias    string `json:"alias"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Alias    string `json:"alias"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    int    `json:"id"`
		Alias string `json:"alias"`
		Name  string `json:"name"`
	} `json:"user"`
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}