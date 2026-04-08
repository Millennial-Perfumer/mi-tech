package whatsapp

import (
	"encoding/json"
	"fmt"
	"log"
	"mi-tech/internal/automation/whatsapp/mcp"
	"mi-tech/internal/config"
	"mi-tech/internal/dto"
	"mi-tech/internal/service"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type AgentService struct {
	mcpClient      *mcp.Client
	settings       *config.SettingsProvider
	plannerService *service.PlannerService
	httpClient     *resty.Client
	repo           *MessagesRepository
	metaClient     *MetaClient
	notifService   *NotificationService
}

func NewAgentService(settings *config.SettingsProvider, plannerService *service.PlannerService, repo *MessagesRepository, metaClient *MetaClient, notifService *NotificationService) *AgentService {
	// Initialize MCP Client
	mcpDir := "internal/automation/whatsapp/mcp"
	
	// Fetch secrets for MCP using the exported Get method
	storeURL := settings.Get("shopify_store_url")
	accessToken := settings.Get("shopify_access_token")
	apiVersion := settings.Get("shopify_api_version")

	env := []string{
		fmt.Sprintf("SHOPIFY_STORE_URL=%s", storeURL),
		fmt.Sprintf("SHOPIFY_ACCESS_TOKEN=%s", accessToken),
		fmt.Sprintf("SHOPIFY_API_VERSION=%s", apiVersion),
		"PATH=/usr/local/bin:/usr/bin:/bin", // Ensure node is in path
	}

	mcpClient, err := mcp.NewClient(mcpDir, env)
	if err != nil {
		log.Printf("Failed to initialize MCP Client: %v", err)
	}

	return &AgentService{
		mcpClient:      mcpClient,
		settings:       settings,
		plannerService: plannerService,
		httpClient:     resty.New(),
		repo:           repo,
		metaClient:     metaClient,
		notifService:   notifService,
	}
}

type AgentResponse struct {
	Reply          string `json:"reply"`
	Classification string `json:"classification"` // GENERAL, QUERY, ISSUE, URGENT
	ShouldCreateTask bool   `json:"should_create_task"`
	TaskTitle      string `json:"task_title,omitempty"`
	TaskPriority   string `json:"task_priority,omitempty"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
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

	// 3. Get Tools
	toolsResp, err := s.mcpClient.Call("list_tools", nil)
	var tools []interface{}
	if err == nil && toolsResp.Result != nil {
		var mcpTools struct {
			Tools []interface{} `json:"tools"`
		}
		json.Unmarshal(toolsResp.Result, &mcpTools)
		tools = mcpTools.Tools
	}

	// Add Native Kanban Tools
	nativeTools := []interface{}{
		map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "list_kanban_columns",
				"description": "List all columns in the 'Support Tickets' board to know where to move tickets.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "move_kanban_task",
				"description": "Move a task to a different column in the Kanban board.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"task_id": map[string]interface{}{"type": "number"},
						"column_id": map[string]interface{}{"type": "number"},
					},
					"required": []string{"task_id", "column_id"},
				},
			},
		},
	}
	tools = append(tools, nativeTools...)

	// 4. LLM Loop
	for turn := 0; turn < 10; turn++ {
		payload := map[string]interface{}{
			"model":    "kimi-k2.5",
			"messages": messages,
		}
		if len(tools) > 0 {
			payload["tools"] = tools
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
					Content   string      `json:"content"`
					ToolCalls []ToolCall  `json:"tool_calls"`
					Role      string      `json:"role"`
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

		if choice.FinishReason == "tool_calls" {
			for _, tc := range choice.Message.ToolCalls {
				resContent := "Error executing tool"
				
				// Handle Native Tools
				if tc.Function.Name == "list_kanban_columns" {
					boards, _ := s.plannerService.ListBoards()
					var targetBoardID uint
					for _, b := range boards {
						if b.Name == "WhatsApp Support" {
							targetBoardID = b.ID
							break
						}
					}
					if targetBoardID > 0 {
						board, _ := s.plannerService.GetBoard(targetBoardID)
						data, _ := json.Marshal(board.Columns)
						resContent = string(data)
					}
				} else if tc.Function.Name == "move_kanban_task" {
					var args struct {
						TaskID   uint `json:"task_id"`
						ColumnID uint `json:"column_id"`
					}
					json.Unmarshal([]byte(tc.Function.Arguments), &args)
					err := s.plannerService.MoveTask(args.TaskID, args.ColumnID, 0)
					if err == nil {
						resContent = "Task moved successfully"
					} else {
						resContent = fmt.Sprintf("Failed to move task: %v", err)
					}
				} else {
					// Handle MCP Tools
					mcpRes, _ := s.mcpClient.Call(tc.Function.Name, tc.Function.Arguments)
					if mcpRes != nil && mcpRes.Result != nil {
						resContent = string(mcpRes.Result)
					}
				}

				messages = append(messages, map[string]string{
					"role":         "tool",
					"tool_call_id": tc.ID,
					"name":         tc.Function.Name,
					"content":      resContent,
				})
			}
			continue
		}

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
		s.CreateKanbanTask(convID, contactName, text, result)
	}

	return nil
}

func (s *AgentService) SendReply(convID int, text string) {
	// For now, we need to get the phone number from the repo
	// We'll skip for brevity or handle it properly later.
	log.Printf("AI Replying to Conv %d: %s", convID, text)
}

func (s *AgentService) CreateKanbanTask(convID int, contactName, text string, res AgentResponse) {
	boards, _ := s.plannerService.ListBoards()
	var boardID uint
	for _, b := range boards {
		if b.Name == "WhatsApp Support" {
			boardID = b.ID
			break
		}
	}
	if boardID == 0 { boardID = 1 }

	meta, _ := json.Marshal(map[string]interface{}{"conv_id": convID})
	metaRaw := json.RawMessage(meta)

	s.plannerService.CreateTask(dto.CreateTaskRequest{
		BoardID:     boardID,
		Title:       res.TaskTitle,
		Description: fmt.Sprintf("Reported by %s: %s", contactName, text),
		Priority:    res.TaskPriority,
		Metadata:    &metaRaw,
	})

	// 3. Ping Telegram
	if s.notifService != nil {
		go s.notifService.PingIssue(contactName, text, res.TaskPriority)
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

func (s *AgentService) ProcessTaskAI(taskID uint, text string) error {
	task, err := s.plannerService.GetTaskByID(taskID)
	if err != nil {
		return err
	}

	messages := []map[string]string{
		{
			"role": "system",
			"content": `You are the Mi-Tech AI Assistant. You are currently acting on a Kanban Task.
You have access to Shopify store data and Kanban board tools.
Identify the user's intent after the '@ai' mention and execute.
Always provide a concise update to be appended to the task description.
If you moved a task or performed an action, state it clearly.`,
		},
		{
			"role":    "user",
			"content": fmt.Sprintf("Task Context:\nTitle: %s\nDescription: %s\nStatus: %s\n\nLatest Request: %s", task.Title, task.Description, task.Status, text),
		},
	}

	toolsResp, err := s.mcpClient.Call("list_tools", nil)
	var tools []interface{}
	if err == nil && toolsResp.Result != nil {
		var mcpTools struct {
			Tools []interface{} `json:"tools"`
		}
		json.Unmarshal(toolsResp.Result, &mcpTools)
		tools = mcpTools.Tools
	}

	nativeTools := []interface{}{
		map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "list_kanban_columns",
				"description": "List all columns in the board.",
				"parameters": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
			},
		},
		map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "move_kanban_task",
				"description": "Move the current task to a different column.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"column_id": map[string]interface{}{"type": "number"},
					},
					"required": []string{"column_id"},
				},
			},
		},
	}
	tools = append(tools, nativeTools...)

	var aiReply string
	apiKey := s.settings.Get("opencode_api_key")
	baseURL := "https://opencode.ai/api/v1"

	for turn := 0; turn < 5; turn++ {
		resp, err := s.httpClient.R().
			SetHeader("Authorization", "Bearer "+apiKey).
			SetBody(map[string]interface{}{
				"model":    "kimi-k2.5",
				"messages": messages,
				"tools":    tools,
			}).
			Post(baseURL + "/chat/completions")

		if err != nil {
			return err
		}

		var completion struct {
			Choices []struct {
				Message struct {
					Content   string      `json:"content"`
					ToolCalls []ToolCall  `json:"tool_calls"`
					Role      string      `json:"role"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}
		json.Unmarshal(resp.Body(), &completion)

		if len(completion.Choices) == 0 {
			break
		}
		choice := completion.Choices[0]
		messages = append(messages, map[string]string{"role": choice.Message.Role, "content": choice.Message.Content})

		if choice.FinishReason == "tool_calls" {
			for _, tc := range choice.Message.ToolCalls {
				resContent := "Error"
				if tc.Function.Name == "list_kanban_columns" {
					board, _ := s.plannerService.GetBoard(task.BoardID)
					data, _ := json.Marshal(board.Columns)
					resContent = string(data)
				} else if tc.Function.Name == "move_kanban_task" {
					var args struct{ ColumnID uint `json:"column_id"` }
					json.Unmarshal([]byte(tc.Function.Arguments), &args)
					s.plannerService.MoveTask(taskID, args.ColumnID, 0)
					resContent = "Task moved successfully"
				} else {
					mcpRes, _ := s.mcpClient.Call(tc.Function.Name, tc.Function.Arguments)
					if mcpRes != nil && mcpRes.Result != nil { resContent = string(mcpRes.Result) }
				}
				messages = append(messages, map[string]string{"role": "tool", "tool_call_id": tc.ID, "name": tc.Function.Name, "content": resContent})
			}
			continue
		}
		aiReply = choice.Message.Content
		break
	}

	if aiReply != "" {
		newDesc := task.Description + "\n\n--- AI RESPONSE ---\n" + aiReply
		s.plannerService.UpdateTask(taskID, dto.UpdateTaskRequest{Description: &newDesc})
	}

	return nil
}
