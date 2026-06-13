package repository

import (
	"mi-tech/internal/domain/planner/entity"
	"strings"
	"time"

	"gorm.io/gorm"
)

type plannerRepository struct {
	db *gorm.DB
}

func NewPlannerRepository(db *gorm.DB) PlannerRepository {
	return &plannerRepository{db: db}
}

// Boards
func (r *plannerRepository) ListBoards() ([]entity.PlannerBoard, error) {
	var boards []entity.PlannerBoard
	err := r.db.Preload("Columns").Find(&boards).Error
	return boards, err
}

func (r *plannerRepository) GetBoardByID(id uint) (entity.PlannerBoard, error) {
	var board entity.PlannerBoard
	err := r.db.Preload("Columns").First(&board, id).Error
	return board, err
}

func (r *plannerRepository) CreateBoard(board *entity.PlannerBoard) error {
	return r.db.Create(board).Error
}

// Columns
func (r *plannerRepository) ListColumns(boardID uint) ([]entity.PlannerColumn, error) {
	var columns []entity.PlannerColumn
	err := r.db.Where("board_id = ?", boardID).Order("\"order\" ASC").Find(&columns).Error
	return columns, err
}

func (r *plannerRepository) UpdateColumnOrder(columns []entity.PlannerColumn) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, col := range columns {
			if err := tx.Model(&entity.PlannerColumn{}).Where("id = ?", col.ID).Update("order", col.Order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Sprints
func (r *plannerRepository) ListSprints(status string) ([]entity.PlannerSprint, error) {
	var sprints []entity.PlannerSprint
	query := r.db
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&sprints).Error
	return sprints, err
}

func (r *plannerRepository) GetSprintByID(id uint) (entity.PlannerSprint, error) {
	var sprint entity.PlannerSprint
	err := r.db.First(&sprint, id).Error
	return sprint, err
}

func (r *plannerRepository) CreateSprint(sprint *entity.PlannerSprint) error {
	return r.db.Create(sprint).Error
}

func (r *plannerRepository) UpdateSprint(sprint *entity.PlannerSprint) error {
	return r.db.Save(sprint).Error
}

func (r *plannerRepository) DeleteSprint(id uint) error {
	return r.db.Delete(&entity.PlannerSprint{}, id).Error
}

func (r *plannerRepository) UpdateSprintStatus(id uint, status string) error {
	return r.db.Model(&entity.PlannerSprint{}).Where("id = ?", id).Update("status", status).Error
}

// Tasks
func (r *plannerRepository) ListTasks(filter PlannerFilter) ([]entity.PlannerTask, error) {
	var tasks []entity.PlannerTask
	query := r.db.Where("board_id = ?", filter.BoardID)

	if filter.SprintID != nil {
		query = query.Where("sprint_id = ?", filter.SprintID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Priority != "" {
		query = query.Where("priority = ?", filter.Priority)
	}
	if filter.Search != "" {
		query = query.Where("title ILIKE ?", "%"+filter.Search+"%")
	}

	err := query.Order("p_order ASC, created_at DESC").Find(&tasks).Error
	return tasks, err
}

func (r *plannerRepository) GetTaskByID(id uint) (entity.PlannerTask, error) {
	var task entity.PlannerTask
	err := r.db.First(&task, id).Error
	return task, err
}

func (r *plannerRepository) CreateTask(task *entity.PlannerTask) error {
	return r.db.Create(task).Error
}

func (r *plannerRepository) UpdateTask(task *entity.PlannerTask) error {
	return r.db.Save(task).Error
}

func (r *plannerRepository) DeleteTask(id uint) error {
	return r.db.Delete(&entity.PlannerTask{}, id).Error
}

func (r *plannerRepository) MoveTask(taskID uint, toColumnID uint, newOrder int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var task entity.PlannerTask
		if err := tx.First(&task, taskID).Error; err != nil {
			return err
		}

		oldColumnID := task.ColumnID
		oldOrder := task.Order

		// 1. Adjust orders in the source column if we are moving out or reordering within
		if oldColumnID != nil {
			if *oldColumnID == toColumnID {
				// Reordering within the same column
				if oldOrder < newOrder {
					// Moving down: decrement order of tasks between old and new
					tx.Model(&entity.PlannerTask{}).
						Where("column_id = ? AND p_order > ? AND p_order <= ?", oldColumnID, oldOrder, newOrder).
						UpdateColumn("p_order", gorm.Expr("p_order - 1"))
				} else if oldOrder > newOrder {
					// Moving up: increment order of tasks between new and old
					tx.Model(&entity.PlannerTask{}).
						Where("column_id = ? AND p_order >= ? AND p_order < ?", oldColumnID, newOrder, oldOrder).
						UpdateColumn("p_order", gorm.Expr("p_order + 1"))
				}
			} else {
				// Moving to a different column: fill the gap in the old column
				tx.Model(&entity.PlannerTask{}).
					Where("column_id = ? AND p_order > ?", oldColumnID, oldOrder).
					UpdateColumn("p_order", gorm.Expr("p_order - 1"))

				// Make space in the new column
				tx.Model(&entity.PlannerTask{}).
					Where("column_id = ? AND p_order >= ?", toColumnID, newOrder).
					UpdateColumn("p_order", gorm.Expr("p_order + 1"))
			}
		} else {
			// Coming from no column (backlog)
			tx.Model(&entity.PlannerTask{}).
				Where("column_id = ? AND p_order >= ?", toColumnID, newOrder).
				UpdateColumn("p_order", gorm.Expr("p_order + 1"))
		}

		var column entity.PlannerColumn
		if err := tx.First(&column, toColumnID).Error; err != nil {
			return err
		}

		// Map status based on column name
		status := "todo"
		colName := strings.ToLower(column.Name)
		if strings.Contains(colName, "progress") || strings.Contains(colName, "doing") {
			status = "in-progress"
		} else if strings.Contains(colName, "done") || strings.Contains(colName, "finish") {
			status = "done"
		} else if strings.Contains(colName, "review") {
			status = "review"
		}

		// Update Task
		updates := map[string]interface{}{
			"column_id":  toColumnID,
			"p_order":    newOrder,
			"status":     status,
			"updated_at": time.Now(),
		}

		if err := tx.Model(&entity.PlannerTask{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
			return err
		}

		// Log Movement
		log := entity.PlannerTaskLog{
			TaskID:       taskID,
			FromColumnID: oldColumnID,
			ToColumnID:   &toColumnID,
			ChangedAt:    time.Now(),
		}
		return tx.Create(&log).Error
	})
}

// Analytics
func (r *plannerRepository) GetSprintVelocity(sprintID uint) (int, error) {
	if sprintID == 0 {
		return 0, nil
	}
	var count int64
	err := r.db.Model(&entity.PlannerTask{}).
		Where("sprint_id = ? AND status = ?", sprintID, "done").
		Count(&count).Error
	return int(count), err
}

func (r *plannerRepository) GetTaskLeadTime(taskID uint) (float64, error) {
	if taskID == 0 {
		return 0, nil
	}
	var task entity.PlannerTask
	if err := r.db.First(&task, taskID).Error; err != nil {
		return 0, err
	}
	if task.CompletedAt == nil {
		return 0, nil
	}
	duration := task.CompletedAt.Sub(task.CreatedAt)
	return duration.Hours() / 24.0, nil // Return in days
}
