package service

import (
	"context"
	"encoding/json"
	"fmt"
	"mi-tech/internal/client/llm"
	"mi-tech/internal/config"
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
)

type AIService struct {
	readRepo    repository.AIReadRepository
	convRepo    repository.AIConversationRepository
	memRepo     repository.AIMemoryRepository
	settings    *config.SettingsProvider
}

func NewAIService(readRepo repository.AIReadRepository, convRepo repository.AIConversationRepository, memRepo repository.AIMemoryRepository, settings *config.SettingsProvider) *AIService {
	return &AIService{
		readRepo:    readRepo,
		convRepo:    convRepo,
		memRepo:     memRepo,
		settings:    settings,
	}
}

func (s *AIService) getProvider() (llm.LLMProvider, error) {
	providerType := s.settings.GetAIProvider()
	enabled := s.settings.IsAIEnabled()
	
	if !enabled {
		return nil, fmt.Errorf("AI analysis is currently disabled in settings")
	}

	if providerType == "local" {
		url := s.settings.GetAILocalURL()
		model := s.settings.GetAILocalModel()
		return llm.NewOllamaProvider(url, model), nil
	}

	apiKey := s.settings.GetOpenAIAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}
	model := s.settings.GetAICloudModel()
	return llm.NewOpenAIProvider(apiKey, model), nil
}

func (s *AIService) getTools() []llm.ToolDef {
	return []llm.ToolDef{
		{
			Name:        "get_revenue_summary",
			Description: "Get total revenue and order count for a specific date range",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"start_date": map[string]interface{}{"type": "string", "description": "ISO date YYYY-MM-DD"},
					"end_date":   map[string]interface{}{"type": "string", "description": "ISO date YYYY-MM-DD"},
				},
				"required": []string{"start_date", "end_date"},
			},
		},
		{
			Name:        "get_revenue_by_channel",
			Description: "Get revenue breakdown by sales channel (Shopify, Amazon, etc.) for a date range",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"start_date": map[string]interface{}{"type": "string", "description": "ISO date YYYY-MM-DD"},
					"end_date":   map[string]interface{}{"type": "string", "description": "ISO date YYYY-MM-DD"},
				},
				"required": []string{"start_date", "end_date"},
			},
		},
		{
			Name:        "get_revenue_by_state",
			Description: "Get revenue breakdown by customer state/region for a date range",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"start_date": map[string]interface{}{"type": "string", "description": "ISO date YYYY-MM-DD"},
					"end_date":   map[string]interface{}{"type": "string", "description": "ISO date YYYY-MM-DD"},
				},
				"required": []string{"start_date", "end_date"},
			},
		},
		{
			Name:        "get_daily_revenue_trend",
			Description: "Get day-by-day revenue for a date range to see trends",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"start_date": map[string]interface{}{"type": "string", "description": "ISO date YYYY-MM-DD"},
					"end_date":   map[string]interface{}{"type": "string", "description": "ISO date YYYY-MM-DD"},
				},
				"required": []string{"start_date", "end_date"},
			},
		},
		{
			Name:        "get_top_products",
			Description: "Get best selling products ranked by quantity sold for a date range. Returns product name, SKU, quantity sold, and revenue.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"start_date": map[string]interface{}{"type": "string", "description": "ISO date YYYY-MM-DD"},
					"end_date":   map[string]interface{}{"type": "string", "description": "ISO date YYYY-MM-DD"},
					"limit":      map[string]interface{}{"type": "integer", "description": "Number of products to return, default 5"},
				},
				"required": []string{"start_date", "end_date"},
			},
		},
		{
			Name:        "get_product_performance",
			Description: "Get detailed stats for a specific product by its SKU — stock levels, total sold, average daily sales",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sku": map[string]interface{}{"type": "string", "description": "Product SKU (e.g. mi-01)"},
				},
				"required": []string{"sku"},
			},
		},
		{
			Name:        "get_customer_segmentation",
			Description: "Get customer analytics: total count, new customers in last 30 days, repeat purchase rate, and channel split",
			Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		{
			Name:        "get_top_customers",
			Description: "Get top customers ranked by total spend",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{"type": "integer", "description": "Number of customers to return, default 10"},
				},
			},
		},
		{
			Name:        "get_inventory_status",
			Description: "Get current inventory stock levels, low stock alerts, and out-of-stock items for all products",
			Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		{
			Name:        "get_business_snapshot",
			Description: "Get a quick overview of today's and this month's key metrics: revenue, orders, low stock count, pending fulfillments",
			Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		{
			Name:        "execute_sql_query",
			Description: "Execute a read-only SQL SELECT query for complex analysis that other tools cannot handle. Use this to join tables, filter data, or perform custom aggregations.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{"type": "string", "description": "The SQL SELECT query to execute. MUST be read-only."},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "list_database_tables",
			Description: "List all available tables in the database to discover what data is available for analysis.",
			Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		{
			Name:        "describe_database_table",
			Description: "Get the column names and data types for a specific table to understand its structure.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"table_name": map[string]interface{}{"type": "string", "description": "The name of the table to describe"},
				},
				"required": []string{"table_name"},
			},
		},
		{
			Name:        "store_business_rule",
			Description: "Save a business rule, logic decision, or user preference to persistent memory so you remember it in future sessions.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key":      map[string]interface{}{"type": "string", "description": "A unique, descriptive slug for this rule (e.g. revenue_exclusion_logic)"},
					"content":  map[string]interface{}{"type": "string", "description": "The full text of the rule or decision"},
					"category": map[string]interface{}{"type": "string", "enum": []string{"business_rule", "user_preference", "analysis_logic"}, "description": "Type of memory"},
				},
				"required": []string{"key", "content"},
			},
		},
		{
			Name:        "list_business_rules",
			Description: "Retrieve all previously saved business rules and decisions from persistent memory.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"category": map[string]interface{}{"type": "string", "description": "Optional category to filter by"},
				},
			},
		},
	}
}

func (s *AIService) executeTool(name string, args string) (string, error) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid tool arguments: %w", err)
	}

	getStr := func(key string) string {
		if v, ok := params[key].(string); ok {
			return v
		}
		return ""
	}

	getInt := func(key string, defaultVal int) int {
		if v, ok := params[key].(float64); ok {
			return int(v)
		}
		return defaultVal
	}

	switch name {
	case "get_revenue_summary":
		res, err := s.readRepo.GetRevenueSummary(getStr("start_date"), getStr("end_date"))
		b, _ := json.Marshal(res)
		return string(b), err
	case "get_revenue_by_channel":
		res, err := s.readRepo.GetRevenueByChannel(getStr("start_date"), getStr("end_date"))
		b, _ := json.Marshal(res)
		return string(b), err
	case "get_revenue_by_state":
		res, err := s.readRepo.GetRevenueByState(getStr("start_date"), getStr("end_date"))
		b, _ := json.Marshal(res)
		return string(b), err
	case "get_daily_revenue_trend":
		res, err := s.readRepo.GetDailyRevenueTrend(getStr("start_date"), getStr("end_date"))
		b, _ := json.Marshal(res)
		return string(b), err
	case "get_top_products":
		limit := getInt("limit", 5)
		res, err := s.readRepo.GetTopProducts(getStr("start_date"), getStr("end_date"), limit)
		b, _ := json.Marshal(res)
		return string(b), err
	case "get_product_performance":
		res, err := s.readRepo.GetProductPerformance(getStr("sku"))
		b, _ := json.Marshal(res)
		return string(b), err
	case "get_customer_segmentation":
		res, err := s.readRepo.GetCustomerSegmentation()
		b, _ := json.Marshal(res)
		return string(b), err
	case "get_top_customers":
		limit := getInt("limit", 10)
		res, err := s.readRepo.GetTopCustomers(limit)
		b, _ := json.Marshal(res)
		return string(b), err
	case "get_inventory_status":
		res, err := s.readRepo.GetInventorySnapshot()
		b, _ := json.Marshal(res)
		return string(b), err
	case "get_business_snapshot":
		res, err := s.readRepo.GetBusinessSnapshot()
		b, _ := json.Marshal(res)
		return string(b), err
	case "execute_sql_query":
		res, err := s.readRepo.ExecuteRawQuery(getStr("query"))
		b, _ := json.Marshal(res)
		return string(b), err
	case "list_database_tables":
		res, err := s.readRepo.ListTables()
		b, _ := json.Marshal(res)
		return string(b), err
	case "describe_database_table":
		res, err := s.readRepo.DescribeTable(getStr("table_name"))
		b, _ := json.Marshal(res)
		return string(b), err
	case "store_business_rule":
		mem := &entity.AIMemory{
			Key:      getStr("key"),
			Content:  getStr("content"),
			Category: getStr("category"),
		}
		if mem.Category == "" {
			mem.Category = "business_rule"
		}
		err := s.memRepo.Upsert(mem)
		return "Successfully saved to persistent memory.", err
	case "list_business_rules":
		res, err := s.memRepo.List(getStr("category"))
		b, _ := json.Marshal(res)
		return string(b), err
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *AIService) Chat(ctx context.Context, userID int64, conversationID int64, userMessage string) (<-chan llm.StreamChunk, error) {
	provider, err := s.getProvider()
	if err != nil {
		return nil, err
	}

	// 1. Persist user message
	if err := s.convRepo.AddMessage(conversationID, "user", userMessage, nil); err != nil {
		return nil, err
	}

	// 2. Load history
	history, err := s.convRepo.GetMessages(conversationID)
	if err != nil {
		return nil, err
	}

	// 3. Load Persistent Memories (Business Rules)
	memories, _ := s.memRepo.List("business_rule")
	rulesText := ""
	for _, m := range memories {
		rulesText += fmt.Sprintf("- %s: %s\n", m.Key, m.Content)
	}
	if rulesText == "" {
		rulesText = "No specific business rules saved yet."
	}

	llmMsgs := []llm.ChatMessage{
		{Role: llm.RoleSystem, Content: fmt.Sprintf(`You are an AI Business Analyst for Millennial Perfumer, a premium D2C fragrance brand.

Your role is to help the business owner understand their data using the tools provided. You have access to specialized aggregate tools, a raw SQL bridge, and database discovery tools.

### 🏛 Established Business Rules (Persistent Memory):
%s

### 🧠 Data Intelligence Strategy:
1. **Discover**: If you are unsure about a table name or its columns, ALWAYS use "list_database_tables" and "describe_database_table" first. Do not guess.
2. **Analyze**: Use the specialized aggregate tools (like get_top_products) for common requests. 
3. **Custom**: Use "execute_sql_query" for complex joins or ad-hoc analysis.
4. **Safety**: "execute_sql_query" is strictly READ-ONLY. Only SELECT queries are allowed.

### High-Level Map (Start Here):
- **orders**: Core sales data (source_id, total_price, status, customer_phone, created_at).
- **order_line_items**: Individual products within orders (sku, title, price, quantity).
- **inventory_items**: Your warehouse stock (mi_sku, current_stock, title).
- **inventory_mappings**: Links internal mi_sku to external platform SKUs (amazon, shopify).
- **customers**: Customer profiles (phone_number, first_name, last_name, total_orders).

### 💡 Data Insights & Common Pitfalls:
- **Customer Linkage**: The "orders" table does NOT have a "customer_id". You must link orders to "customers" using "orders.customer_phone" vs "customers.phone_number". Both are normalized (e.g. +91...), but use TRIM() for robustness.
- **Status Variations**: Order statuses vary by source.
  - Shopify: 'paid', 'fulfilled', 'CANCELLED' (uppercase).
  - Amazon: 'Shipped', 'Unshipped', 'Canceled' (mixed case).
  - Rule: When filtering non-cancelled orders, ALWAYS use LOWER(status) NOT IN ('cancelled', 'canceled').
- **Calculated Stats**: The "customers" table has "total_orders" and "total_spent" columns. These are useful for quick checks, but use the "orders" table for date-specific analysis.
- **Amazon Mappings**: To analyze Amazon sales, join "order_line_items" (on sku) to "inventory_mappings" (on external_sku where platform='amazon') to find the internal "mi_sku".

### Guidelines:
- **Conciseness**: Give direct, data-driven answers.
- **Visuals**: Use tables for comparisons and lists for trends.
- **Professionalism**: Never mention technical terms like "SQL", "GORM", or internal table names in your final answer.
- **Product Names**: Always prioritize showing the product TITLE over the SKU code.
- **Date Ranges**: When asked for "today" or "this month", determine the date range relative to the current local time.
`, rulesText)},
	}
	for _, m := range history {
		msg := llm.ChatMessage{Role: llm.ChatRole(m.Role), Content: m.Content}
		if m.Metadata != nil {
			json.Unmarshal(*m.Metadata, &msg)
		}
		llmMsgs = append(llmMsgs, msg)
	}

	outCh := make(chan llm.StreamChunk)

	go func() {
		defer close(outCh)

		// This loop handles multi-step tool calling
		for i := 0; i < 10; i++ { // Limit to 10 turns to prevent loops
			stream, err := provider.StreamChat(ctx, llmMsgs, s.getTools())
			if err != nil {
				outCh <- llm.StreamChunk{Error: err}
				return
			}

			var currentContent string
			var currentToolCalls []llm.ToolCall
			var isToolTurn bool

			for chunk := range stream {
				if chunk.Error != nil {
					outCh <- chunk
					return
				}

				if len(chunk.ToolCalls) > 0 {
					isToolTurn = true
					// Accumulate tool calls
					for _, tc := range chunk.ToolCalls {
						if tc.ID != "" {
							currentToolCalls = append(currentToolCalls, tc)
						} else if len(currentToolCalls) > 0 {
							// Append to last tool call arguments
							idx := len(currentToolCalls) - 1
							currentToolCalls[idx].Function.Arguments += tc.Function.Arguments
						}
					}
				}

				if chunk.Content != "" {
					currentContent += chunk.Content
					// Only stream content to user if we haven't detected tool calls yet
					if !isToolTurn {
						outCh <- llm.StreamChunk{Content: chunk.Content}
					}
				}

				if chunk.Done {
					break
				}
			}

			if !isToolTurn {
				// Final response reached
				outCh <- llm.StreamChunk{Done: true}
				
				// Persist assistant message
				s.convRepo.AddMessage(conversationID, "assistant", currentContent, nil)
				return
			}

			// Handle Tool Calls
			assistantMsg := llm.ChatMessage{
				Role:      llm.RoleAssistant,
				Content:   currentContent,
				ToolCalls: currentToolCalls,
			}
			llmMsgs = append(llmMsgs, assistantMsg)
			
			// Persist assistant tool call message
			meta, _ := json.Marshal(assistantMsg)
			rawMeta := json.RawMessage(meta)
			s.convRepo.AddMessage(conversationID, "assistant", currentContent, &rawMeta)
			
			// Execute tools
			for _, tc := range currentToolCalls {
				result, err := s.executeTool(tc.Function.Name, tc.Function.Arguments)
				if err != nil {
					result = fmt.Sprintf("Error: %v", err)
				}
				
				// Add tool result to history
				toolMsg := llm.ChatMessage{
					Role:       llm.RoleTool,
					Content:    result,
					ToolCallID: &tc.ID,
				}
				llmMsgs = append(llmMsgs, toolMsg)
				
				// Persist tool message
				toolMeta, _ := json.Marshal(toolMsg)
				rawToolMeta := json.RawMessage(toolMeta)
				s.convRepo.AddMessage(conversationID, "tool", result, &rawToolMeta)
			}
			// Loop continues for the next turn
		}
		
		outCh <- llm.StreamChunk{Error: fmt.Errorf("AI reached maximum reasoning depth")}
	}()

	return outCh, nil
}

func (s *AIService) CreateConversation(userID int64, firstMessage string) (*entity.AIConversation, error) {
	title := firstMessage
	if len(title) > 40 {
		title = title[:37] + "..."
	}
	return s.convRepo.CreateConversation(userID, title)
}

func (s *AIService) ListConversations(userID int64) ([]entity.AIConversation, error) {
	return s.convRepo.ListConversations(userID, 50)
}

func (s *AIService) GetConversation(id, userID int64) (map[string]interface{}, error) {
	conv, err := s.convRepo.GetConversation(id, userID)
	if err != nil {
		return nil, err
	}
	messages, err := s.convRepo.GetMessages(id)
	if err != nil {
		return nil, err
	}

	// Filter internal messages for the frontend (tool calls, etc)
	filtered := make([]entity.AIMessage, 0)
	for _, m := range messages {
		if m.Role == "tool" {
			continue
		}
		// Skip assistant messages that only have tool calls and no content
		if m.Role == "assistant" && m.Content == "" && m.Metadata != nil {
			continue
		}
		filtered = append(filtered, m)
	}

	return map[string]interface{}{
		"conversation": conv,
		"messages":     filtered,
	}, nil
}

func (s *AIService) DeleteConversation(id, userID int64) error {
	return s.convRepo.DeleteConversation(id, userID)
}
