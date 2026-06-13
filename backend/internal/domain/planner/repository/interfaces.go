package repository

import (
	"mi-tech/internal/domain/planner/entity"
)

// PlannerFilter holds query parameters for tasks and analytics.
type PlannerFilter struct {
	BoardID  uint
	SprintID *uint
	Status   string
	Priority string
	Search   string
}

// PlannerRepository defines all data access operations for the planner module.
type PlannerRepository interface {
	// Boards
	ListBoards() ([]entity.PlannerBoard, error)
	GetBoardByID(id uint) (entity.PlannerBoard, error)
	CreateBoard(board *entity.PlannerBoard) error

	// Columns
	ListColumns(boardID uint) ([]entity.PlannerColumn, error)
	UpdateColumnOrder(columns []entity.PlannerColumn) error

	// Sprints
	ListSprints(status string) ([]entity.PlannerSprint, error)
	GetSprintByID(id uint) (entity.PlannerSprint, error)
	CreateSprint(sprint *entity.PlannerSprint) error
	UpdateSprint(sprint *entity.PlannerSprint) error
	DeleteSprint(id uint) error
	UpdateSprintStatus(id uint, status string) error

	// Tasks
	ListTasks(filter PlannerFilter) ([]entity.PlannerTask, error)
	GetTaskByID(id uint) (entity.PlannerTask, error)
	CreateTask(task *entity.PlannerTask) error
	UpdateTask(task *entity.PlannerTask) error
	DeleteTask(id uint) error
	MoveTask(taskID uint, toColumnID uint, newOrder int) error

	// Analytics
	GetSprintVelocity(sprintID uint) (int, error)
	GetTaskLeadTime(taskID uint) (float64, error)
}
