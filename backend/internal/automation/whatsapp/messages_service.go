package whatsapp

import (
	"mi-tech/internal/config"
	"time"
)

type MessagesService struct {
	repo       *MessagesRepository
	metaClient *MetaClient
	settings   *config.SettingsProvider
}

func NewMessagesService(repo *MessagesRepository, settings *config.SettingsProvider) *MessagesService {
	metaClient := NewMetaClient(settings)
	return &MessagesService{
		repo:       repo,
		metaClient: metaClient,
		settings:   settings,
	}
}

func (s *MessagesService) SendTemplateMessage(storeID string, templateID int, orderID int64, phoneNumber, templateName, languageCode string, components []interface{}) error {
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
		SentAt:      time.Now().UTC(),
	})

	return err
}

func (s *MessagesService) HandleStatusUpdate(messageID, status string) error {
	return s.repo.UpdateMessageStatus(messageID, status)
}

func (s *MessagesService) GetMessages(storeID string, startDate, endDate *time.Time, search string, templateName string, limit, offset int) ([]AutomationMessage, error) {
	return s.repo.GetMessages(storeID, startDate, endDate, search, templateName, limit, offset)
}

func (s *MessagesService) GetActiveTemplateNamesForFilter(storeID string, startDate, endDate *time.Time, search string) ([]string, error) {
	return s.repo.GetActiveTemplateNamesForFilter(storeID, startDate, endDate, search)
}

func (s *MessagesService) GetMessagesCount(storeID string, startDate, endDate *time.Time, search string, templateName string) (int, error) {
	return s.repo.GetMessagesCount(storeID, startDate, endDate, search, templateName)
}

func (s *MessagesService) GetAutomationMetrics(storeID string, startDate, endDate *time.Time) (map[string]interface{}, error) {
	return s.repo.GetAutomationMetrics(storeID, startDate, endDate)
}

func (s *MessagesService) SyncMetricsFromMeta(startDate, endDate string) (map[string]interface{}, error) {
	analytics, err := s.metaClient.GetTemplateAnalytics(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// For local metrics, we need to treat the strings as IST and convert to UTC bounds
	// but since we're in the service layer, we might not have the helper.
	// Actually, the easiest is to just pass the strings if the repo still supports them,
	// or have the handler pass the parsed times.
	// Let's assume the repo handles the conversion if we pass it *time.Time.

	// I'll keep it simple for now and update the repo to handle *time.Time.
	// We'll trust the caller (handler) to pass parsed times if they have them,
	// or we can parse them here using a similar logic.

	// However, SyncMetricsFromMeta is called with strings. Let's parse them.
	ist := time.FixedZone("IST", 5*3600+1800)
	var start, end *time.Time
	if t, err := time.ParseInLocation("2006-01-02", startDate, ist); err == nil {
		utc := t.UTC()
		start = &utc
	}
	if t, err := time.ParseInLocation("2006-01-02", endDate, ist); err == nil {
		utc := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, ist).UTC()
		end = &utc
	}

	localMetrics, _ := s.repo.GetAutomationMetrics("", start, end)
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
