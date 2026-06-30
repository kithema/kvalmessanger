package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"2say/internal/middleware"
	"2say/internal/models"
)

func (h Handler) CreateChat(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	log.Printf("=== CREATE CHAT START ===")
	log.Printf("userID from context: %d", userID)
	
	if userID == 0 {
		log.Printf("ERROR: userID is 0, unauthorized")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Проверяем, существует ли пользователь
	var exists bool
	err := h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		log.Printf("ERROR checking user existence: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		log.Printf("ERROR: User with ID %d does not exist in database", userID)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	log.Printf("User with ID %d exists", userID)

	var req models.CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR decoding request: %v", err)
		http.Error(w, "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Request data: aliases=%v, isGroup=%v, name=%s", req.UserAliases, req.IsGroup, req.Name)

	if len(req.UserAliases) == 0 {
		log.Printf("ERROR: No participants provided")
		http.Error(w, "At least one participant is required", http.StatusBadRequest)
		return
	}

	// Получаем ID пользователей по их alias
	var userIDs []int
	for _, alias := range req.UserAliases {
		alias = strings.TrimSpace(alias)
		if alias == "" {
			continue
		}
		var id int
		err := h.DB.QueryRow("SELECT id FROM users WHERE alias = $1", alias).Scan(&id)
		if err != nil {
			log.Printf("ERROR: User with alias '%s' not found: %v", alias, err)
			http.Error(w, "User with alias '"+alias+"' not found", http.StatusBadRequest)
			return
		}
		log.Printf("Found user: alias=%s, id=%d", alias, id)
		
		if id == userID {
			log.Printf("ERROR: User tried to add himself: %d", id)
			http.Error(w, "You cannot add yourself as a participant", http.StatusBadRequest)
			return
		}
		userIDs = append(userIDs, id)
	}

	log.Printf("All found user IDs: %v", userIDs)

	if len(userIDs) == 0 {
		log.Printf("ERROR: No valid participants found")
		http.Error(w, "No valid participants found", http.StatusBadRequest)
		return
	}

	// Если это личный чат, проверяем, не существует ли уже
	if !req.IsGroup && len(userIDs) == 1 {
		log.Printf("Checking if private chat already exists between user %d and %d", userID, userIDs[0])
		var existingID int
		err := h.DB.QueryRow(`
			SELECT c.id FROM chats c
			JOIN chat_members cp1 ON c.id = cp1.chat_id
			JOIN chat_members cp2 ON c.id = cp2.chat_id
			WHERE c.is_group = false 
			AND cp1.user_id = $1 
			AND cp2.user_id = $2
		`, userID, userIDs[0]).Scan(&existingID)

		if err == nil {
			log.Printf("Private chat already exists with ID: %d", existingID)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]int{"chat_id": existingID})
			return
		}
		log.Printf("No existing chat found, creating new one")
	}

	// Создаем новый чат
	log.Printf("Starting transaction to create chat")
	tx, err := h.DB.Begin()
	if err != nil {
		log.Printf("ERROR starting transaction: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Генерируем имя для чата
	chatName := req.Name
	if chatName == "" {
		if req.IsGroup {
			var names []string
			for _, uid := range userIDs {
				var name string
				err := tx.QueryRow("SELECT COALESCE(name, alias) FROM users WHERE id = $1", uid).Scan(&name)
				if err == nil && name != "" {
					names = append(names, name)
				}
			}
			if len(names) > 0 {
				chatName = strings.Join(names, ", ")
				if len(chatName) > 30 {
					chatName = chatName[:27] + "..."
				}
			} else {
				chatName = "Group Chat"
			}
		} else {
			var otherName string
			err := tx.QueryRow("SELECT COALESCE(name, alias) FROM users WHERE id = $1", userIDs[0]).Scan(&otherName)
			if err == nil {
				chatName = otherName
			} else {
				chatName = "Private Chat"
			}
		}
	}

	log.Printf("Creating chat with name: %s", chatName)

	// Вставляем чат с правильными типами
	var chatID int
	now := time.Now()
	err = tx.QueryRow(
		"INSERT INTO chats (name, is_group, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id",
		chatName, req.IsGroup, now, now,
	).Scan(&chatID)
	if err != nil {
		log.Printf("ERROR creating chat: %v", err)
		http.Error(w, "Cannot create chat", http.StatusInternalServerError)
		return
	}

	log.Printf("Chat created with ID: %d", chatID)

	// Добавляем всех участников (включая создателя)
	allUserIDs := append([]int{userID}, userIDs...)
	log.Printf("Adding participants: %v", allUserIDs)
	
	for _, uid := range allUserIDs {
		_, err = tx.Exec(
			"INSERT INTO chat_members (chat_id, user_id, joined_at) VALUES ($1, $2, $3)",
			chatID, uid, now,
		)
		if err != nil {
			log.Printf("ERROR adding participant %d: %v", uid, err)
			http.Error(w, "Cannot add participants", http.StatusInternalServerError)
			return
		}
		log.Printf("Added participant: user_id=%d", uid)
	}

	if err = tx.Commit(); err != nil {
		log.Printf("ERROR committing transaction: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	log.Printf("=== CHAT CREATED SUCCESSFULLY: ID=%d ===", chatID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"chat_id": chatID})
}

func (h Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		log.Printf("ERROR: Unauthorized send message attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR decoding send message request: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	log.Printf("SendMessage: userID=%d, chatID=%d, text=%s", userID, req.ChatID, req.Text)

	// Проверяем, является ли пользователь участником чата
	var isParticipant bool
	err := h.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM chat_members WHERE chat_id=$1 AND user_id=$2)",
		req.ChatID, userID,
	).Scan(&isParticipant)
	if err != nil || !isParticipant {
		log.Printf("ERROR: User %d is not a participant of chat %d", userID, req.ChatID)
		http.Error(w, "You are not a participant of this chat", http.StatusForbidden)
		return
	}

	var messageID int
	err = h.DB.QueryRow(`
		INSERT INTO messages (chat_id, sender_id, reply_to_id, text, sent_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`, req.ChatID, userID, req.ReplyToID, req.Text, time.Now()).Scan(&messageID)
	if err != nil {
		log.Printf("ERROR inserting message: %v", err)
		http.Error(w, "Cannot send message", http.StatusInternalServerError)
		return
	}

	// Обновляем время чата
	_, err = h.DB.Exec(
		"UPDATE chats SET updated_at = $1 WHERE id = $2",
		time.Now(), req.ChatID,
	)
	if err != nil {
		log.Printf("WARNING: Could not update chat timestamp: %v", err)
	}

	log.Printf("Message sent successfully: ID=%d", messageID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"message_id": messageID})
}

func (h Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	chatIDStr := r.URL.Query().Get("chat_id")
	if chatIDStr == "" {
		http.Error(w, "chat_id required", http.StatusBadRequest)
		return
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
		return
	}

	log.Printf("GetMessages: userID=%d, chatID=%d", userID, chatID)

	// Проверяем участие в чате
	var isParticipant bool
	err = h.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM chat_members WHERE chat_id=$1 AND user_id=$2)",
		chatID, userID,
	).Scan(&isParticipant)
	if err != nil || !isParticipant {
		log.Printf("ERROR: User %d is not a participant of chat %d", userID, chatID)
		http.Error(w, "You are not a participant of this chat", http.StatusForbidden)
		return
	}

	limit := 50
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	rows, err := h.DB.Query(`
		SELECT m.id, m.chat_id, m.sender_id, m.reply_to_id, m.text, 
		       m.sent_at, m.edited_at, m.is_deleted, m.photo_url, m.pdf_url, 
		       u.alias, u.name
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.chat_id = $1 AND m.is_deleted = false
		ORDER BY m.sent_at DESC
		LIMIT $2 OFFSET $3
	`, chatID, limit, offset)
	if err != nil {
		log.Printf("ERROR getting messages: %v", err)
		http.Error(w, "Cannot get messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var id, chatID, senderID int
		var replyToID *int
		var text string
		var sentAt time.Time
		var editedAt *time.Time
		var isDeleted bool
		var photoURL, pdfURL *string
		var senderAlias, senderName string

		err := rows.Scan(
			&id, &chatID, &senderID, &replyToID,
			&text, &sentAt, &editedAt, &isDeleted,
			&photoURL, &pdfURL, &senderAlias, &senderName,
		)
		if err != nil {
			log.Printf("ERROR scanning message row: %v", err)
			continue
		}

		msg := map[string]interface{}{
			"id":           id,
			"chat_id":      chatID,
			"sender_id":    senderID,
			"sender_alias": senderAlias,
			"sender_name":  senderName,
			"text":         text,
			"sent_at":      sentAt,
			"is_deleted":   isDeleted,
		}
		if replyToID != nil {
			msg["reply_to_id"] = *replyToID
		}
		if editedAt != nil {
			msg["edited_at"] = *editedAt
		}
		if photoURL != nil {
			msg["photo_url"] = *photoURL
		}
		if pdfURL != nil {
			msg["pdf_url"] = *pdfURL
		}
		messages = append(messages, msg)
	}

	log.Printf("Returning %d messages for chat %d", len(messages), chatID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (h Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	messageIDStr := r.URL.Query().Get("message_id")
	if messageIDStr == "" {
		http.Error(w, "message_id required", http.StatusBadRequest)
		return
	}

	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		http.Error(w, "Invalid message_id", http.StatusBadRequest)
		return
	}

	log.Printf("DeleteMessage: userID=%d, messageID=%d", userID, messageID)

	// Проверяем, что пользователь является отправителем
	var senderID int
	err = h.DB.QueryRow(
		"SELECT sender_id FROM messages WHERE id = $1",
		messageID,
	).Scan(&senderID)
	if err != nil {
		log.Printf("ERROR message not found: %v", err)
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	if senderID != userID {
		log.Printf("ERROR: User %d tried to delete message %d owned by %d", userID, messageID, senderID)
		http.Error(w, "You can only delete your own messages", http.StatusForbidden)
		return
	}

	_, err = h.DB.Exec(
		"UPDATE messages SET is_deleted = true, edited_at = $1 WHERE id = $2",
		time.Now(), messageID,
	)
	if err != nil {
		log.Printf("ERROR deleting message: %v", err)
		http.Error(w, "Cannot delete message", http.StatusInternalServerError)
		return
	}

	log.Printf("Message %d deleted successfully", messageID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h Handler) GetUserChats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("GetUserChats: userID=%d", userID)

	// Простой запрос без подсчета непрочитанных сообщений
	query := `
		SELECT 
			c.id, 
			c.name, 
			COALESCE(c.is_group, false) as is_group,
			c.photo_url,
			c.updated_at
		FROM chats c
		INNER JOIN chat_members cm ON c.id = cm.chat_id
		WHERE cm.user_id = $1
		ORDER BY c.updated_at DESC
	`

	rows, err := h.DB.Query(query, userID)
	if err != nil {
		log.Printf("ERROR getting user chats: %v", err)
		http.Error(w, "Cannot get chats: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var chats []map[string]interface{}
	for rows.Next() {
		var id int
		var name *string
		var isGroup bool
		var photoURL *string
		var updatedAt time.Time

		err := rows.Scan(&id, &name, &isGroup, &photoURL, &updatedAt)
		if err != nil {
			log.Printf("ERROR scanning chat row: %v", err)
			continue
		}

		chatName := ""
		if name != nil {
			chatName = *name
		}
		
		// Для личных чатов получаем alias собеседника
		if !isGroup && chatName == "" {
			var otherAlias string
			err := h.DB.QueryRow(`
				SELECT u.alias 
				FROM users u
				INNER JOIN chat_members cm ON u.id = cm.user_id
				WHERE cm.chat_id = $1 AND cm.user_id != $2
				LIMIT 1
			`, id, userID).Scan(&otherAlias)
			if err == nil {
				chatName = otherAlias
			} else {
				chatName = "Private Chat"
			}
		}

		chat := map[string]interface{}{
			"id":           id,
			"name":         chatName,
			"is_group":     isGroup,
			"photo_url":    photoURL,
			"updated_at":   updatedAt,
			"unread_count": 0, // Пока просто 0
		}
		chats = append(chats, chat)
	}

	log.Printf("Returning %d chats for user %d", len(chats), userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chats)
}

func (h Handler) GetChatMembers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	chatIDStr := r.URL.Query().Get("chat_id")
	if chatIDStr == "" {
		http.Error(w, "chat_id required", http.StatusBadRequest)
		return
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
		return
	}

	// Проверяем, является ли пользователь участником чата
	var isParticipant bool
	err = h.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM chat_members WHERE chat_id=$1 AND user_id=$2)",
		chatID, userID,
	).Scan(&isParticipant)
	if err != nil || !isParticipant {
		http.Error(w, "You are not a participant of this chat", http.StatusForbidden)
		return
	}

	rows, err := h.DB.Query(`
		SELECT u.id, u.alias, u.name
		FROM users u
		JOIN chat_members cm ON u.id = cm.user_id
		WHERE cm.chat_id = $1
	`, chatID)
	if err != nil {
		http.Error(w, "Cannot get members", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var members []map[string]interface{}
	for rows.Next() {
		var id int
		var alias, name string
		err := rows.Scan(&id, &alias, &name)
		if err != nil {
			continue
		}
		members = append(members, map[string]interface{}{
			"user_id": id,
			"alias":   alias,
			"name":    name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}