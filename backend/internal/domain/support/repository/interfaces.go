package repository

import (
	"mi-tech/internal/domain/support/entity"
)

type TicketRepository interface {
	Create(ticket *entity.SupportTicket) error
	GetByID(id uint) (*entity.SupportTicket, error)
	GetByTicketID(ticketID string) (*entity.SupportTicket, error)
	List() ([]entity.SupportTicket, error)
	Update(ticket *entity.SupportTicket) error
	Delete(id uint) error
	GetNextTicketNumber() (string, error)
}
