package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mi-tech/internal/domain/communication/entity"
	repository "mi-tech/internal/domain/communication/repository"
	customerRepoPkg "mi-tech/internal/domain/order/repository"
	"mi-tech/internal/shared/config"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MessagesService struct {
	repo         repository.MessagesRepository
	metaClient   *MetaClient
	settings     *config.SettingsProvider
	customerRepo *customerRepoPkg.CustomerRepository
	agentService *AgentService
}

func NewMessagesService(repo repository.MessagesRepository, settings *config.SettingsProvider, customerRepo *customerRepoPkg.CustomerRepository, agentService *AgentService) *MessagesService {
	metaClient := NewMetaClient(settings)
	s := &MessagesService{
		repo:         repo,
		metaClient:   metaClient,
		settings:     settings,
		customerRepo: customerRepo,
		agentService: agentService,
	}

	// Trigger early cleanup and start daily ticker
	go func() {
		s.CleanupOldMedia()
		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			s.CleanupOldMedia()
		}
	}()

	return s
}

func (s *MessagesService) SendTemplateMessage(storeID string, templateID int, orderID int64, phoneNumber, templateName, languageCode string, components []interface{}) error {
	// 1. Send via Meta
	msgID, err := s.metaClient.SendTemplateMessage(phoneNumber, templateName, languageCode, components)
	if err != nil {
		// Log failed attempt
		s.repo.SaveMessage(entity.AutomationMessage{
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
	_, err = s.repo.SaveMessage(entity.AutomationMessage{
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

func (s *MessagesService) GetMessages(storeID string, startDate, endDate *time.Time, search string, templateName string, limit, offset int) ([]entity.AutomationMessage, error) {
	return s.repo.GetMessages(storeID, startDate, endDate, search, templateName, limit, offset)
}

func (s *MessagesService) GetMessagesByOrderID(orderID int64) ([]entity.AutomationMessage, error) {
	return s.repo.GetMessagesByOrderID(orderID)
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

func (s *MessagesService) GetConversations() ([]entity.Conversation, error) {
	return s.repo.GetConversations()
}

func (s *MessagesService) GetChatMessages(conversationID int, limit, offset int) ([]entity.ChatMessage, error) {
	return s.repo.GetChatMessages(conversationID, limit, offset)
}

func (s *MessagesService) UpdateConversationMode(id int, mode string) error {
	return s.repo.UpdateConversationMode(id, mode)
}

func (s *MessagesService) SendMediaMessage(phoneNumber, mediaID, mediaType, caption string, senderRole string) (int, error) {
	// 1. Send via Meta API
	msgID, err := s.metaClient.SendMediaMessage(phoneNumber, mediaID, mediaType, caption)
	if err != nil {
		return 0, err
	}

	// 2. Upsert conversation
	displayText := fmt.Sprintf("Sent a %s", mediaType)
	if caption != "" {
		displayText = caption
	}
	convID, err := s.repo.UpsertConversation(phoneNumber, "", displayText)
	if err != nil {
		return 0, err
	}

	// 3. Save chat message
	metadata := map[string]interface{}{
		"media_id": mediaID,
		"caption":  caption,
	}
	metadataBytes, _ := json.Marshal(metadata)

	chatMsg := entity.ChatMessage{
		ConversationID: convID,
		MessageID:      msgID,
		Text:           displayText,
		Type:           mediaType,
		Direction:      "outgoing",
		SenderRole:     senderRole,
		Status:         "sent",
		SentAt:         time.Now().UTC(),
		Metadata:       metadataBytes,
	}
	return s.repo.SaveChatMessage(chatMsg)
}

func (s *MessagesService) SendFreeTextMessage(phoneNumber, text string, senderRole string) (int, error) {
	// 1. Send via Meta API
	msgID, err := s.metaClient.SendTextMessage(phoneNumber, text)
	if err != nil {
		return 0, err
	}

	// 2. Upsert conversation
	convID, err := s.repo.UpsertConversation(phoneNumber, "", text)
	if err != nil {
		return 0, err
	}

	// 3. Save chat message
	chatMsg := entity.ChatMessage{
		ConversationID: convID,
		MessageID:      msgID,
		Text:           text,
		Type:           "text",
		Direction:      "outgoing",
		SenderRole:     senderRole,
		Status:         "sent",
		SentAt:         time.Now().UTC(),
	}
	return s.repo.SaveChatMessage(chatMsg)
}

func (s *MessagesService) HandleIncomingMessage(phoneNumber, contactName, messageID, text, msgType string, metadata []byte) error {
	// 1. Check if customer exists in our DB to get their preferred name
	if s.customerRepo != nil {
		// Using a background context for the lookup; we can refine this if a request-scoped context is available.
		if cust, err := s.customerRepo.GetByPhone(context.Background(), phoneNumber); err == nil && cust != nil {
			var dbName string
			if cust.FirstName != nil && *cust.FirstName != "" {
				dbName = *cust.FirstName
			}
			if cust.LastName != nil && *cust.LastName != "" {
				if dbName != "" {
					dbName += " "
				}
				dbName += *cust.LastName
			}

			// If we found a name in the DB, override the profile name from WhatsApp
			if dbName != "" {
				contactName = dbName
			}
		}
	}

	// 2. Upsert conversation
	convID, err := s.repo.UpsertConversation(phoneNumber, contactName, text)
	if err != nil {
		return err
	}

	// 2. Save chat message
	chatMsg := entity.ChatMessage{
		ConversationID: convID,
		MessageID:      messageID,
		Text:           text,
		Type:           msgType,
		Direction:      "incoming",
		SenderRole:     "user",
		Status:         "received",
		SentAt:         time.Now().UTC(),
		Metadata:       metadata,
	}
	_, err = s.repo.SaveChatMessage(chatMsg)
	if err != nil {
		return err
	}

	// 3. Send automated support redirect if this is not a reaction
	if msgType != "reaction" {
		replyText := "Hi! This channel is used only for updates.\n\nFor support, please chat with our team here:\n👉 https://wa.me/917904769823"

		go func() {
			_, err := s.SendFreeTextMessage(phoneNumber, replyText, "system")
			if err != nil {
				log.Printf("Failed to send support redirect to %s: %v", phoneNumber, err)
			}
		}()
	}

	// 4. Trigger Agent Service if mode is auto
	conv, _ := s.repo.GetConversationByPhone(phoneNumber)
	if conv != nil && conv.Mode == "auto" && s.agentService != nil {
		// Run in background to not block webhook
		go func() {
			if err := s.agentService.ProcessMessage(convID, contactName, text); err != nil {
				log.Printf("Agent Service Error: %v", err)
			}
		}()
	}

	return nil
}

func (s *MessagesService) StoreMedia(mediaID string, data []byte, contentType string) (string, error) {
	// 1. Determine extension
	ext := ".bin"
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "image/jpeg") || strings.Contains(ct, "image/jpg") {
		ext = ".jpg"
	} else if strings.Contains(ct, "image/png") {
		ext = ".png"
	} else if strings.Contains(ct, "image/webp") || strings.Contains(ct, "webp") {
		ext = ".webp"
	} else if strings.Contains(ct, "image/gif") {
		ext = ".gif"
	} else if strings.Contains(ct, "video/mp4") {
		ext = ".mp4"
	} else if strings.Contains(ct, "video/quicktime") {
		ext = ".mov"
	} else if strings.Contains(ct, "audio/mpeg") || strings.Contains(ct, "audio/mp3") {
		ext = ".mp3"
	} else if strings.Contains(ct, "audio/ogg") {
		ext = ".ogg"
	} else if strings.Contains(ct, "application/pdf") {
		ext = ".pdf"
	} else if strings.Contains(ct, "application/msword") || strings.Contains(ct, "application/vnd.openxmlformats-officedocument.wordprocessingml.document") {
		ext = ".docx"
	}

	filename := mediaID + ext
	dir := filepath.Join("uploads", "whatsapp")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	fullPath := filepath.Join(dir, filename)

	// 2. Save to disk
	err := os.WriteFile(fullPath, data, 0644)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (s *MessagesService) DownloadAndStoreMedia(mediaID string) (string, error) {
	// 1. Get media URL from Meta
	url, err := s.metaClient.GetMediaURL(mediaID)
	if err != nil {
		return "", err
	}

	// 2. Download bytes
	data, contentType, err := s.metaClient.DownloadMedia(url)
	if err != nil {
		return "", err
	}

	// 3. Store locally
	return s.StoreMedia(mediaID, data, contentType)
}

func (s *MessagesService) CleanupOldMedia() {
	dir := filepath.Join("uploads", "whatsapp")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Media Cleanup Error: Failed to read directory: %v", err)
		return
	}

	count := 0
	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			continue
		}
		// Policy: 15 days
		if time.Since(info.ModTime()) > 15*24*time.Hour {
			os.Remove(filepath.Join(dir, f.Name()))
			count++
		}
	}
	if count > 0 {
		log.Printf("Media Cleanup: Removed %d old WhatsApp media files", count)
	}
}

func (s *MessagesService) GetMetaClient() *MetaClient {
	return s.metaClient
}

func (s *MessagesService) GetRepo() repository.MessagesRepository {
	return s.repo
}
