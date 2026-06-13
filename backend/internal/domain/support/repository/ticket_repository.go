package repository

import (
	"fmt"
	"mi-tech/internal/domain/support/entity"

	"gorm.io/gorm"
)

type ticketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) Create(ticket *entity.SupportTicket) error {
	return r.db.Create(ticket).Error
}

func (r *ticketRepository) GetByID(id uint) (*entity.SupportTicket, error) {
	var ticket entity.SupportTicket
	err := r.db.First(&ticket, id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) GetByTicketID(ticketID string) (*entity.SupportTicket, error) {
	var ticket entity.SupportTicket
	err := r.db.Where("ticket_id = ?", ticketID).First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) List() ([]entity.SupportTicket, error) {
	var tickets []entity.SupportTicket
	err := r.db.Order("created_at DESC").Find(&tickets).Error
	return tickets, err
}

func (r *ticketRepository) Update(ticket *entity.SupportTicket) error {
	return r.db.Save(ticket).Error
}

func (r *ticketRepository) Delete(id uint) error {
	return r.db.Delete(&entity.SupportTicket{}, id).Error
}

func (r *ticketRepository) GetNextTicketNumber() (string, error) {
	var count int64
	err := r.db.Model(&entity.SupportTicket{}).Count(&count).Error
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("TIC-%d", 1001+count), nil
}
