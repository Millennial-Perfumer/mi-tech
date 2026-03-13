package whatsapp

import (
	"mi-tech/internal/config"
	"time"
)

type MessagesService struct {
	repo       *MessagesRepository
	metaClient *MetaClient
	cfg        *config.Config
}

func NewMessagesService(repo *MessagesRepository, cfg *config.Config) *MessagesService {
	metaClient := NewMetaClient(cfg.WhatsAppAccessToken, cfg.WhatsAppPhoneNumberID, cfg.WhatsAppWABAID)
	return &MessagesService{
		repo:       repo,
		metaClient: metaClient,
		cfg:        cfg,
	}
}

func (s *MessagesService) SendTemplateMessage(storeID string, templateID int, orderID, phoneNumber, templateName, languageCode string, components []interface{}) error {
	// 1. Send via Meta
	msgID, err := s.metaClient.SendTemplateMessage(phoneNumber, templateName, languageCode, components)
	if err != nil {
		// Log failed attempt
		s.repo.SaveMessage(AutomationMessage{
			StoreID:      storeID,
			TemplateID:   templateID,
			OrderID:      orderID,
			PhoneNumber:  phoneNumber,
			Status:       "failed",
			ErrorMessage: err.Error(),
		})
		return err
	}

	// 2. record success
	_, err = s.repo.SaveMessage(AutomationMessage{
		StoreID:     storeID,
		TemplateID:  templateID,
		OrderID:     orderID,
		PhoneNumber: phoneNumber,
		MessageID:   msgID,
		Status:      "sent",
		SentAt:      time.Now(),
	})

	return err
}

func (s *MessagesService) HandleStatusUpdate(messageID, status string) error {
	return s.repo.UpdateMessageStatus(messageID, status)
}

func (s *MessagesService) GetMessages(storeID string) ([]AutomationMessage, error) {
	return s.repo.GetMessages(storeID)
}

func (s *MessagesService) GetAutomationMetrics(storeID string) (map[string]interface{}, error) {
	return s.repo.GetAutomationMetrics(storeID)
}
