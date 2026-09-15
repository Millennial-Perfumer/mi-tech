package service

import (
	"encoding/json"
	"fmt"
	"log"
	repository "mi-tech/internal/domain/communication/repository"
	supportDto "mi-tech/internal/domain/support/dto"
	supportServicePkg "mi-tech/internal/domain/support/service"
	"mi-tech/internal/shared/config"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type AgentService struct {
	settings       *config.SettingsProvider
	supportService *supportServicePkg.TicketService
	httpClient     *resty.Client
	repo           repository.MessagesRepository
	metaClient     *MetaClient
	NotifService   *NotificationService
}

func NewNewAgentService(
	settings *config.SettingsProvider,
	supportService *supportServicePkg.TicketService,
	repo repository.MessagesRepository,
	metaClient *MetaClient,
	notifService *NotificationService,
) *AgentService {
	return &AgentService{
		settings:       settings,
		supportService: supportService,
		httpClient:     resty.New(),
		repo:           repo,
		metaClient:     metaClient,
		NotifService:   notifService,
	}
}

type AgentResponse struct {
	Reply            string `json:"reply"`
	Classification   string `json:"classification"` // GENERAL, QUERY, ISSUE, URGENT
	ShouldCreateTask bool   `json:"should_create_task"`
	TaskTitle        string `json:"task_title,omitempty"`
	TaskPriority     string `json:"task_priority,omitempty"`
}

func (s *AgentService) ProcessMessage(convID int, contactName, text string) error {
	var result AgentResponse

	// 1. Fetch conversation history for context
	history, err := s.repo.GetChatMessages(convID, 10, 0)
	if err != nil {
		return err
	}

	// 2. Prepare the prompt
	apiKey := s.settings.Get("opencode_api_key")
	if apiKey == "" {
		return fmt.Errorf("missing opencode_api_key in settings")
	}

	baseURL := "https://opencode.ai/api/v1"

	messages := []map[string]string{
		{
			"role": "system",
			"content": `You are the Mi-Tech AI Assistant. You handle customer queries via WhatsApp.
You have access to Shopify store data via tools.

Classification Rules:
- ISSUE: Broken product, incorrect order, delayed shipment.
- URGENT: Angry customer, critical failure.
- QUERY: Searching for products, stock, or status.

Output MUST be a JSON object: 
{
  "reply": "friendly response",
  "classification": "GENERAL|QUERY|ISSUE|URGENT",
  "should_create_task": boolean,
  "task_title": "Short title describing the ticket",
  "task_priority": "low|medium|high|urgent"
}
NOTE: When should_create_task is true, it will generate a formal support ticket.`,
		},
	}

	for _, m := range history {
		role := "assistant"
		if m.Direction == "incoming" {
			role = "user"
		}
		messages = append(messages, map[string]string{
			"role":    role,
			"content": m.Text,
		})
	}

	// 3. LLM Loop
	for turn := 0; turn < 10; turn++ {
		payload := map[string]interface{}{
			"model":    "kimi-k2.5",
			"messages": messages,
		}
		resp, err := s.httpClient.R().
			SetHeader("Authorization", "Bearer "+apiKey).
			SetBody(payload).
			Post(baseURL + "/chat/completions")

		if err != nil {
			return err
		}

		var completion struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
					Role    string `json:"role"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}
		json.Unmarshal(resp.Body(), &completion)

		if len(completion.Choices) == 0 {
			break
		}

		choice := completion.Choices[0]
		msg := map[string]string{
			"role":    choice.Message.Role,
			"content": choice.Message.Content,
		}
		messages = append(messages, msg)

		// Parse Final JSON
		if err := json.Unmarshal([]byte(choice.Message.Content), &result); err != nil {
			// Extract JSON from markdown if needed
			if strings.Contains(choice.Message.Content, "{") {
				start := strings.Index(choice.Message.Content, "{")
				end := strings.LastIndex(choice.Message.Content, "}")
				if start >= 0 && end > start {
					json.Unmarshal([]byte(choice.Message.Content[start:end+1]), &result)
				}
			}
			if result.Reply == "" {
				result.Reply = choice.Message.Content
			}
		}
		break
	}

	// 5. Execution
	if result.Reply != "" {
		s.SendReply(convID, result.Reply)
	}

	if result.ShouldCreateTask {
		s.CreateSupportTicket(convID, contactName, text, result)
	}

	return nil
}

func (s *AgentService) SendReply(convID int, text string) {
	log.Printf("AI Replying to Conv %d: %s", convID, text)
}

func (s *AgentService) CreateSupportTicket(convID int, contactName, text string, res AgentResponse) {
	s.supportService.CreateTicket(supportDto.CreateTicketRequest{
		Title:       res.TaskTitle,
		Description: fmt.Sprintf("Reported by %s: %s (conv_id: %d)", contactName, text, convID),
		Priority:    res.TaskPriority,
	})

	// 3. Ping Telegram
	if s.NotifService != nil {
		go s.NotifService.PingIssue(contactName, text, res.TaskPriority)
	}
}

func (s *AgentService) GenerateDailyConcernsSummary() (string, error) {
	// Fetch issues from the last 24 hours
	since := time.Now().Add(-24 * time.Hour)
	issues, err := s.repo.GetIssuesSince(since)
	if err != nil {
		return "", err
	}

	if len(issues) == 0 {
		return "No new concerns detected in the last 24 hours. Smooth sailing! 🚢", nil
	}

	// Prepare data for LLM
	var content strings.Builder
	content.WriteString("Analyze the following customer concerns from today and provide a concise summary for the admin. Organize by priority.\n\n")
	for _, issue := range issues {
		content.WriteString(fmt.Sprintf("[%s] %s\n", issue.Priority, issue.Text))
	}

	messages := []map[string]string{
		{
			"role":    "system",
			"content": "You are a senior support manager. Summarize the following concerns into a professional executive summary for the admin. Use bullet points and highlight critical issues. End with a pulse check (Overall sentiment).",
		},
		{
			"role":    "user",
			"content": content.String(),
		},
	}

	apiKey := s.settings.Get("opencode_api_key")
	baseURL := "https://opencode.ai/api/v1"

	resp, err := s.httpClient.R().
		SetHeader("Authorization", "Bearer "+apiKey).
		SetBody(map[string]interface{}{
			"model":    "kimi-k2.5",
			"messages": messages,
		}).
		Post(baseURL + "/chat/completions")

	if err != nil {
		return "", err
	}

	var completion struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(resp.Body(), &completion)

	if len(completion.Choices) > 0 {
		return completion.Choices[0].Message.Content, nil
	}

	return "Failed to generate summary.", fmt.Errorf("empty response from LLM")
}
