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

func (s *MessagesService) GetMessages(storeID string, startDate, endDate string, limit, offset int) ([]AutomationMessage, error) {
	return s.repo.GetMessages(storeID, startDate, endDate, limit, offset)
}

func (s *MessagesService) GetMessagesCount(storeID string, startDate, endDate string) (int, error) {
	return s.repo.GetMessagesCount(storeID, startDate, endDate)
}

func (s *MessagesService) GetAutomationMetrics(storeID string, startDate, endDate string) (map[string]interface{}, error) {
	return s.repo.GetAutomationMetrics(storeID, startDate, endDate)
}

func (s *MessagesService) SyncMetricsFromMeta(startDate, endDate string) (map[string]interface{}, error) {
	analytics, err := s.metaClient.GetTemplateAnalytics(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Consolidate: Fetch local metrics once instead of multiple redundant calls
	localMetrics, _ := s.repo.GetAutomationMetrics("", startDate, endDate)
	triggered := localMetrics["triggered"].(int)
	failed := localMetrics["failed"].(int)

	var totalSent, totalDelivered, totalRead int
	for _, t := range analytics {
		totalSent += t.SentCount
		totalDelivered += t.DeliveredCount
		totalRead += t.ReadCount
	}

	readRate := 0.0
	if totalSent > 0 {
		readRate = (float64(totalRead) / float64(totalSent)) * 100
	}

	return map[string]interface{}{
		"sent":      totalSent,
		"delivered": totalDelivered,
		"read":      totalRead,
		"read_rate": readRate,
		"triggered": triggered,
		"failed":    failed,
	}, nil
}
