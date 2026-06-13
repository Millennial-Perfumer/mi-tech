package test

import (
	"testing"

	"mi-tech/internal/domain/support/dto"
	"mi-tech/internal/domain/support/entity"
	"mi-tech/internal/domain/support/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTicketRepository struct {
	mock.Mock
}

func (m *mockTicketRepository) Create(ticket *entity.SupportTicket) error {
	args := m.Called(ticket)
	return args.Error(0)
}

func (m *mockTicketRepository) GetByID(id uint) (*entity.SupportTicket, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SupportTicket), args.Error(1)
}

func (m *mockTicketRepository) GetByTicketID(ticketID string) (*entity.SupportTicket, error) {
	args := m.Called(ticketID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SupportTicket), args.Error(1)
}

func (m *mockTicketRepository) List() ([]entity.SupportTicket, error) {
	args := m.Called()
	return args.Get(0).([]entity.SupportTicket), args.Error(1)
}

func (m *mockTicketRepository) Update(ticket *entity.SupportTicket) error {
	args := m.Called(ticket)
	return args.Error(0)
}

func (m *mockTicketRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockTicketRepository) GetNextTicketNumber() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestTicketService_CreateTicket(t *testing.T) {
	mockRepo := new(mockTicketRepository)
	srv := service.NewTicketService(mockRepo)

	mockRepo.On("GetNextTicketNumber").Return("TIC-1001", nil)
	mockRepo.On("Create", mock.Anything).Return(nil)

	req := dto.CreateTicketRequest{
		Title:       "Test Ticket",
		Description: "Test Desc",
		Priority:    "high",
	}

	res, err := srv.CreateTicket(req)
	assert.NoError(t, err)
	assert.Equal(t, "TIC-1001", res.TicketID)
	assert.Equal(t, "Test Ticket", res.Title)
	assert.Equal(t, "Test Desc", res.Description)
	assert.Equal(t, "high", res.Priority)
	assert.Equal(t, "open", res.Status)

	mockRepo.AssertExpectations(t)
}
