package entity

import (
	"time"
)

type SupportTicket struct {
	ID          uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TicketID    string    `gorm:"column:ticket_id;unique;not null" json:"ticket_id"`
	Title       string    `gorm:"column:title;not null" json:"title"`
	Description string    `gorm:"column:description" json:"description"`
	Priority    string    `gorm:"column:priority;default:medium" json:"priority"` // low, medium, high, urgent
	Status      string    `gorm:"column:status;default:open" json:"status"`       // open, in-progress, resolved, closed
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SupportTicket) TableName() string { return "support_tickets" }
