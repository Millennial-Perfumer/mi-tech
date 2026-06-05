package handler

import (
	"encoding/json"
	"fmt"
	"mi-tech/internal/service"
	"net/http"
	"strconv"
)

type AIHandler struct {
	aiService *service.AIService
}

func NewAIHandler(aiService *service.AIService) *AIHandler {
	return &AIHandler{aiService: aiService}
}

func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ConversationID int64  `json:"conversation_id"`
		Message        string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value("userID").(int64)
	if userID == 0 {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	// If no conversation ID, create a new one
	if req.ConversationID == 0 {
		conv, err := h.aiService.CreateConversation(userID, req.Message)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.ConversationID = conv.ID
	}

	// Set up SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")

	// Flush the conversation ID first so the client knows it
	fmt.Fprintf(w, "data: {\"conversation_id\": %d}\n\n", req.ConversationID)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	ctx := r.Context()
	stream, err := h.aiService.Chat(ctx, userID, req.ConversationID, req.Message)
	if err != nil {
		fmt.Fprintf(w, "data: {\"error\": %q}\n\n", err.Error())
		return
	}

	for chunk := range stream {
		if chunk.Error != nil {
			fmt.Fprintf(w, "data: {\"error\": %q}\n\n", chunk.Error.Error())
			return
		}

		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", string(data))

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func (h *AIHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("userID").(int64)

	convs, err := h.aiService.ListConversations(userID) // Need to implement this in service
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(convs)
}

func (h *AIHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	userID, _ := r.Context().Value("userID").(int64)

	conv, err := h.aiService.GetConversation(id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(conv)
}

func (h *AIHandler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	userID, _ := r.Context().Value("userID").(int64)

	err := h.aiService.DeleteConversation(id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
