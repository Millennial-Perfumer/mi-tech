package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"mi-tech/internal/domain/user/dto"
	"mi-tech/internal/domain/user/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetUsers returns all registered users.
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	log.Printf("UserHandler.GetUsers: received request")
	users, err := h.userService.GetUsers()
	if err != nil {
		log.Printf("UserHandler.GetUsers: error fetching users: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("UserHandler.GetUsers: found %d users", len(users))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"users":   users,
	})
}

// CreateUser creates a new user.
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" || req.Role == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	if err := h.userService.CreateUser(req.Username, req.Password, req.Role); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "user created successfully",
	})
}
