package models

import (
	"time"
)

type Chat struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	IsGroup   bool      `json:"is_group"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	PhotoURL  string    `json:"photo_url"`
}

type ChatParticipant struct {
	ChatID          int       `json:"chat_id"`
	UserID          int       `json:"user_id"`
	LastReadMessage int       `json:"last_read_message"`
	JoinedAt        time.Time `json:"joined_at"`
}

type Message struct {
	ID        int       `json:"id"`
	ChatID    int       `json:"chat_id"`
	SenderID  int       `json:"sender_id"`
	ReplyToID *int      `json:"reply_to_id,omitempty"`
	Text      string    `json:"text"`
	SentAt    time.Time `json:"sent_at"`
	EditedAt  *time.Time `json:"edited_at,omitempty"`
	IsDeleted bool      `json:"is_deleted"`
	PhotoURL  *string   `json:"photo_url,omitempty"`
	PDFURL    *string   `json:"pdf_url,omitempty"`
}

type SendMessageRequest struct {
	ChatID    int    `json:"chat_id"`
	Text      string `json:"text"`
	ReplyToID *int   `json:"reply_to_id,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
	PDFURL    string `json:"pdf_url,omitempty"`
}

type CreateChatRequest struct {
	Name        string   `json:"name"`
	IsGroup     bool     `json:"is_group"`
	UserIDs     []int    `json:"user_ids,omitempty"`
	UserAliases []string `json:"user_aliases"`
}

type UpdateProfileRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}