package service

import (
	"fmt"
	"time"

	"mi-tech/internal/domain/support/dto"
	"mi-tech/internal/domain/support/entity"
	"mi-tech/internal/domain/support/repository"
)

type TicketService struct {
	repo repository.TicketRepository
}

func NewTicketService(repo repository.TicketRepository) *TicketService {
	return &TicketService{repo: repo}
}

func (s *TicketService) CreateTicket(req dto.CreateTicketRequest) (dto.TicketResponse, error) {
	ticketID, err := s.repo.GetNextTicketNumber()
	if err != nil {
		return dto.TicketResponse{}, fmt.Errorf("failed to generate ticket ID: %w", err)
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	ticket := &entity.SupportTicket{
		TicketID:    ticketID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    priority,
		Status:      "open",
	}

	if err := s.repo.Create(ticket); err != nil {
		return dto.TicketResponse{}, fmt.Errorf("failed to create ticket: %w", err)
	}

	return toResponse(ticket), nil
}

func (s *TicketService) ListTickets() ([]dto.TicketResponse, error) {
	tickets, err := s.repo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list tickets: %w", err)
	}

	res := make([]dto.TicketResponse, len(tickets))
	for i, t := range tickets {
		res[i] = toResponse(&t)
	}
	return res, nil
}

func (s *TicketService) GetTicket(id uint) (dto.TicketResponse, error) {
	ticket, err := s.repo.GetByID(id)
	if err != nil {
		return dto.TicketResponse{}, fmt.Errorf("failed to find ticket: %w", err)
	}
	return toResponse(ticket), nil
}

func (s *TicketService) UpdateTicketStatus(id uint, status string) (dto.TicketResponse, error) {
	ticket, err := s.repo.GetByID(id)
	if err != nil {
		return dto.TicketResponse{}, fmt.Errorf("failed to find ticket: %w", err)
	}

	ticket.Status = status
	ticket.UpdatedAt = time.Now()

	if err := s.repo.Update(ticket); err != nil {
		return dto.TicketResponse{}, fmt.Errorf("failed to update ticket status: %w", err)
	}

	return toResponse(ticket), nil
}

func toResponse(t *entity.SupportTicket) dto.TicketResponse {
	return dto.TicketResponse{
		ID:          t.ID,
		TicketID:    t.TicketID,
		Title:       t.Title,
		Description: t.Description,
		Priority:    t.Priority,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
