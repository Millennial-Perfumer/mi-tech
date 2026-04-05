package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"mi-tech/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// GetUsers returns all registered users.
// @Summary List users
// @Description Retrieve a list of all users. Required 'admin' role.
// @Tags users
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /users [get]
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
// @Summary Create user
// @Description Register a new user with a specified role. Required 'admin' role.
// @Tags users
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body CreateUserRequest true "User data"
// @Success 200 {object} map[string]interface{}
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
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
