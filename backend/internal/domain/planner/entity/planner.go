package entity

import (
	"encoding/json"
	"time"
)

// TaskMetadata is the structured representation of the metadata JSON field.
type TaskMetadata struct {
	Subtasks []Subtask `json:"subtasks,omitempty"`
}

// Subtask represents a single item in a task's checklist.
type Subtask struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// PlannerBoard represents the "planner_boards" table.
type PlannerBoard struct {
	ID          uint            `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string          `gorm:"column:name;not null" json:"name"`
	Description string          `gorm:"column:description" json:"description"`
	CreatedAt   time.Time       `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
	Columns     []PlannerColumn `gorm:"foreignKey:BoardID" json:"columns,omitempty"`
}

func (PlannerBoard) TableName() string { return "planner_boards" }

// PlannerColumn represents the "planner_columns" table.
type PlannerColumn struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	BoardID   uint      `gorm:"column:board_id;not null" json:"board_id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Order     int       `gorm:"column:order;not null" json:"order"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (PlannerColumn) TableName() string { return "planner_columns" }

// PlannerSprint represents the "planner_sprints" table.
type PlannerSprint struct {
	ID        uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string     `gorm:"column:name;not null" json:"name"`
	Goal      string     `gorm:"column:goal" json:"goal"`
	StartDate *time.Time `gorm:"column:start_date" json:"start_date"`
	EndDate   *time.Time `gorm:"column:end_date" json:"end_date"`
	Status    string     `gorm:"column:status;default:planned" json:"status"` // planned, active, completed
	CreatedAt time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (PlannerSprint) TableName() string { return "planner_sprints" }

// PlannerTask represents the "planner_tasks" table.
type PlannerTask struct {
	ID          uint             `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	BoardID     uint             `gorm:"column:board_id;not null" json:"board_id"`
	ColumnID    *uint            `gorm:"column:column_id" json:"column_id"`
	SprintID    *uint            `gorm:"column:sprint_id" json:"sprint_id"`
	Title       string           `gorm:"column:title;not null" json:"title"`
	Description string           `gorm:"column:description" json:"description"`
	Priority    string           `gorm:"column:priority;default:medium" json:"priority"` // low, medium, high, urgent
	Status      string           `gorm:"column:status;default:todo" json:"status"`       // todo, in-progress, done, archived
	Metadata    *json.RawMessage `gorm:"column:metadata;type:jsonb" json:"metadata"`
	CompletedAt *time.Time       `gorm:"column:completed_at" json:"completed_at"`
	CreatedAt   time.Time        `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time        `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
	Order       int              `gorm:"column:p_order;default:0" json:"order"`
}

func (PlannerTask) TableName() string { return "planner_tasks" }

// PlannerTaskLog represents the "planner_task_logs" table.
type PlannerTaskLog struct {
	ID           uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TaskID       uint      `gorm:"column:task_id;not null" json:"task_id"`
	FromColumnID *uint     `gorm:"column:from_column_id" json:"from_column_id"`
	ToColumnID   *uint     `gorm:"column:to_column_id" json:"to_column_id"`
	ChangedAt    time.Time `gorm:"column:changed_at;default:CURRENT_TIMESTAMP" json:"changed_at"`
}

func (PlannerTaskLog) TableName() string { return "planner_task_logs" }
