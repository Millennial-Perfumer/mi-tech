package handler

import (
	"encoding/json"
	"mi-tech/internal/automation/whatsapp"
	"net/http"
	"strconv"
	"strings"

	"mi-tech/internal/dto"
	"mi-tech/internal/repository"
	"mi-tech/internal/service"
)

type PlannerHandler struct {
	service      *service.PlannerService
	agentService *whatsapp.AgentService
}

func NewPlannerHandler(service *service.PlannerService, agentService *whatsapp.AgentService) *PlannerHandler {
	return &PlannerHandler{service: service, agentService: agentService}
}

// GetBoards handles GET /api/planner/boards
func (h *PlannerHandler) GetBoards(w http.ResponseWriter, r *http.Request) {
	boards, err := h.service.ListBoards()
	if err != nil {
		http.Error(w, "Failed to fetch boards", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"boards":  boards,
	})
}

// GetTasks handles GET /api/planner/tasks
func (h *PlannerHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	boardID, _ := strconv.Atoi(r.URL.Query().Get("board_id"))
	sprintIDStr := r.URL.Query().Get("sprint_id")
	status := r.URL.Query().Get("status")
	priority := r.URL.Query().Get("priority")
	search := r.URL.Query().Get("search")

	filter := repository.PlannerFilter{
		BoardID:  uint(boardID),
		Status:   status,
		Priority: priority,
		Search:   search,
	}
	if sprintIDStr != "" {
		sid, _ := strconv.Atoi(sprintIDStr)
		uid := uint(sid)
		filter.SprintID = &uid
	}

	tasks, err := h.service.ListTasks(filter)
	if err != nil {
		http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tasks":   tasks,
	})
}

// CreateTask handles POST /api/planner/tasks
func (h *PlannerHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	task, err := h.service.CreateTask(req)
	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"task":    task,
	})
}

// MoveTask handles POST /api/planner/tasks/move
func (h *PlannerHandler) MoveTask(w http.ResponseWriter, r *http.Request) {
	var req dto.MoveTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	taskID := req.TaskID
	if taskID == 0 {
		// Fallback to query param for backward compatibility if needed,
		// but frontend is already sending task_id in body.
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.Atoi(idStr)
		taskID = uint(id)
	}

	if taskID == 0 {
		http.Error(w, "Missing task_id", http.StatusBadRequest)
		return
	}

	if req.ToColumnID == 0 {
		http.Error(w, "Missing to_column_id (target column not identified)", http.StatusBadRequest)
		return
	}

	if err := h.service.MoveTask(taskID, req.ToColumnID, req.NewOrder); err != nil {
		http.Error(w, "Failed to move task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch updated task to return it
	task, _ := h.service.GetTaskByID(taskID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"task":    task,
	})
}

// UpdateTask handles PUT /api/planner/tasks
func (h *PlannerHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req dto.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	task, err := h.service.UpdateTask(uint(id), req)
	if err != nil {
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	// AI Mention Hook
	if req.Description != nil && strings.Contains(*req.Description, "@ai") {
		go h.agentService.ProcessTaskAI(uint(id), *req.Description)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"task":    task,
	})
}

// DeleteTask handles DELETE /api/planner/tasks
func (h *PlannerHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteTask(uint(id)); err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// GetSprints handles GET /api/planner/sprints
func (h *PlannerHandler) GetSprints(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	sprints, err := h.service.ListSprints(status)
	if err != nil {
		http.Error(w, "Failed to fetch sprints", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"sprints": sprints,
	})
}

// CreateSprint handles POST /api/planner/sprints
func (h *PlannerHandler) CreateSprint(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSprintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	sprint, err := h.service.CreateSprint(req)
	if err != nil {
		http.Error(w, "Failed to create sprint", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"sprint":  sprint,
	})
}

func (h *PlannerHandler) UpdateSprint(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req dto.UpdateSprintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	sprint, err := h.service.UpdateSprint(uint(id), req)
	if err != nil {
		http.Error(w, "Failed to update sprint", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"sprint":  sprint,
	})
}

func (h *PlannerHandler) DeleteSprint(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteSprint(uint(id)); err != nil {
		http.Error(w, "Failed to delete sprint", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// GetAnalytics handles GET /api/planner/analytics
func (h *PlannerHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	sprintID, _ := strconv.Atoi(r.URL.Query().Get("sprint_id"))
	taskID, _ := strconv.Atoi(r.URL.Query().Get("task_id"))

	analytics, err := h.service.GetAnalytics(uint(sprintID), uint(taskID))
	if err != nil {
		// Log error but don't 500 if it's just 'not found' for empty params
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"analytics": dto.PlannerAnalyticsResponse{SprintVelocity: 0, TaskLeadTime: 0},
			"message":   "Aggregate view",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"analytics": analytics,
	})
}
