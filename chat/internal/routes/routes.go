package routes

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"2say/internal/handlers"
	"2say/internal/middleware"
)

func SetupRoutes(db *sql.DB) *chi.Mux {
	r := chi.NewRouter()

	// CORS для фронтенда
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://127.0.0.1:5500"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.Logger)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	h := &handlers.Handler{DB: db}

	// Публичные маршруты
	r.Post("/api/register", h.RegisterHandler)
	r.Post("/api/login", h.LoginHandler)

	// Защищенные маршруты
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Get("/profile", h.GetProfile)
		r.Put("/profile", h.UpdateProfile)

		r.Get("/chats", h.GetUserChats)
		r.Post("/chats", h.CreateChat)

		r.Post("/messages", h.SendMessage)
		r.Get("/messages", h.GetMessages)
		r.Delete("/messages", h.DeleteMessage)

		r.Get("/chats/members", h.GetChatMembers)
	})

	return r
}