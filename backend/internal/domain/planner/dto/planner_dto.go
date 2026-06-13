package dto

import (
	"encoding/json"
	"mi-tech/internal/domain/planner/entity"
	"time"
)

type PlannerBoardResponse struct {
	ID          uint                    `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	Columns     []PlannerColumnResponse `json:"columns,omitempty"`
}

type PlannerColumnResponse struct {
	ID      uint   `json:"id"`
	BoardID uint   `json:"board_id"`
	Name    string `json:"name"`
	Order   int    `json:"order"`
}

type PlannerSprintResponse struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Goal      string     `json:"goal"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Status    string     `json:"status"`
}

type PlannerTaskResponse struct {
	ID          uint             `json:"id"`
	BoardID     uint             `json:"board_id"`
	ColumnID    *uint            `json:"column_id"`
	SprintID    *uint            `json:"sprint_id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Priority    string           `json:"priority"`
	Status      string           `json:"status"`
	Metadata    *json.RawMessage `json:"metadata"`
	CompletedAt *time.Time       `json:"completed_at"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Order       int              `json:"order"`
}

type CreateTaskRequest struct {
	BoardID     uint             `json:"board_id"`
	ColumnID    *uint            `json:"column_id"`
	SprintID    *uint            `json:"sprint_id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Priority    string           `json:"priority"`
	Metadata    *json.RawMessage `json:"metadata"`
}

type MoveTaskRequest struct {
	TaskID     uint `json:"task_id"`
	ToColumnID uint `json:"to_column_id"`
	NewOrder   int  `json:"new_order"`
}

type CreateSprintRequest struct {
	Name      string    `json:"name"`
	Goal      string    `json:"goal"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

type UpdateSprintRequest struct {
	Name      *string    `json:"name,omitempty"`
	Goal      *string    `json:"goal,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Status    *string    `json:"status,omitempty"`
}

type UpdateTaskRequest struct {
	Title       *string          `json:"title,omitempty"`
	Description *string          `json:"description,omitempty"`
	Priority    *string          `json:"priority,omitempty"`
	Status      *string          `json:"status,omitempty"`
	SprintID    *uint            `json:"sprint_id,omitempty"`
	Metadata    *json.RawMessage `json:"metadata,omitempty"`
}

type PlannerAnalyticsResponse struct {
	SprintVelocity int     `json:"sprint_velocity"`
	TaskLeadTime   float64 `json:"task_lead_time_days"`
}

func PlannerBoardToResponse(e entity.PlannerBoard) PlannerBoardResponse {
	cols := make([]PlannerColumnResponse, len(e.Columns))
	for i, c := range e.Columns {
		cols[i] = PlannerColumnResponse{
			ID:      c.ID,
			BoardID: c.BoardID,
			Name:    c.Name,
			Order:   c.Order,
		}
	}
	return PlannerBoardResponse{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
		Columns:     cols,
	}
}

func PlannerTaskToResponse(e entity.PlannerTask) PlannerTaskResponse {
	return PlannerTaskResponse{
		ID:          e.ID,
		BoardID:     e.BoardID,
		ColumnID:    e.ColumnID,
		SprintID:    e.SprintID,
		Title:       e.Title,
		Description: e.Description,
		Priority:    e.Priority,
		Status:      e.Status,
		Metadata:    e.Metadata,
		CompletedAt: e.CompletedAt,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
		Order:       e.Order,
	}
}
