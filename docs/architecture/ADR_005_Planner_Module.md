# ADR-005: Planner Module Architecture

## Context and Problem Statement
The application requires a project management layer to track internal tasks and sprint progress. Currently, no task tracking exists. The goal is a clean, minimalistic Kanban board with sprint planning and performance analytics.

## Decision Drivers
- **Minimalism**: Low visual noise, high focus.
- **Performance**: Instant UI feedback during drag-and-drop.
- **Traceability**: Linkage to existing orders/customers.
- **Analytics**: Ability to derive velocity and lead times.

## Proposed Architecture

### 1. Unified Task-Board-Sprint Relationship
We will use a multi-level hierarchy:
- **Board**: A logical grouping of work.
- **Column**: Status-based stages on a board.
- **Sprint**: Time-boxed windows for task execution.
- **Task**: The central unit of work, belonging to a board and optionally a sprint.

### 2. Data Model (GORM Entities)

```go
type PlannerBoard struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Columns     []PlannerColumn `gorm:"foreignKey:BoardID" json:"columns"`
}

type PlannerColumn struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	BoardID uint   `gorm:"index" json:"board_id"`
	Name    string `gorm:"not null" json:"name"`
	Order   int    `gorm:"not null" json:"order"`
}

type PlannerSprint struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Goal      string    `json:"goal"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Status    string    `gorm:"default:'planned'" json:"status"` // planned, active, completed
}

type PlannerTask struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	BoardID     uint           `gorm:"index" json:"board_id"`
	ColumnID    uint           `gorm:"index" json:"column_id"`
	SprintID    *uint          `gorm:"index" json:"sprint_id"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `json:"description"`
	Priority    string         `gorm:"default:'medium'" json:"priority"` // low, medium, high, urgent
	Status      string         `gorm:"default:'todo'" json:"status"`     // todo, in-progress, done, archived
	Metadata    JSON           `gorm:"type:jsonb" json:"metadata"`       // { "order_id": 123, "customer_id": 456 }
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
```

### 3. State Management (Task Movement)
To ensure high-performance analytics, state changes will be logged in a `planner_task_logs` table. This allows us to calculate time-in-column and velocity without querying the main task table recursively.

### 4. Interface Philosophy
- **Frontend**: Use `@dnd-kit/core` for accessible, performant drag-and-drop.
- **Theming**: Zinc (900) backgrounds, Slate (400) text, and Voltage Blue (500) for primary highlights.
- **REST Consistency**: Follow existing backend handler patterns for JSON responses and error handling.

## Consequences
- **Positive**: Enables advanced reporting (Burn-down charts, Velocity).
- **Negative**: Adds database complexity with task logging.
- **Mitigation**: Use Postgres JSONB for flexible task metadata to avoid frequent schema changes for different task types.
