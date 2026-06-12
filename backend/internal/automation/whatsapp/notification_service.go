package whatsapp

import (
	"fmt"
	"mi-tech/internal/config"
	"strings"

	"github.com/go-resty/resty/v2"
)

type NotificationService struct {
	settings   *config.SettingsProvider
	httpClient *resty.Client
}

func NewNotificationService(settings *config.SettingsProvider) *NotificationService {
	return &NotificationService{
		settings:   settings,
		httpClient: resty.New(),
	}
}

func (s *NotificationService) PingIssue(contactName, text, priority string) error {
	token := s.settings.Get("telegram_bot_token")
	chatID := s.settings.Get("telegram_chat_id")

	if token == "" || chatID == "" {
		return fmt.Errorf("telegram configuration missing")
	}

	message := fmt.Sprintf("🚨 *New WhatsApp Issue Detected!*\n\n"+
		"*From:* %s\n"+
		"*Priority:* %s\n"+
		"*Message:* %s\n\n"+
		"Check the Kanban board for details.",
		contactName, strings.ToUpper(priority), text)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	resp, err := s.httpClient.R().
		SetFormData(map[string]string{
			"chat_id":    chatID,
			"text":       message,
			"parse_mode": "Markdown",
		}).
		Post(url)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("telegram api error: %s", resp.String())
	}

	return nil
}

func (s *NotificationService) SendSummary(summary string) error {
	token := s.settings.Get("telegram_bot_token")
	chatID := s.settings.Get("telegram_chat_id")

	if token == "" || chatID == "" {
		return fmt.Errorf("telegram configuration missing")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	_, err := s.httpClient.R().
		SetFormData(map[string]string{
			"chat_id":    chatID,
			"text":       summary,
			"parse_mode": "Markdown",
		}).
		Post(url)

	return err
}
