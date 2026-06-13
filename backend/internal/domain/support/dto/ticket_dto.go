package dto

import "time"

type CreateTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"` // low, medium, high, urgent
}

type UpdateTicketStatusRequest struct {
	Status string `json:"status"` // open, in-progress, resolved, closed
}

type TicketResponse struct {
	ID          uint      `json:"id"`
	TicketID    string    `json:"ticket_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
