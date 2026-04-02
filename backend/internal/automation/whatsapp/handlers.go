package whatsapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mi-tech/internal/config"
	"mi-tech/internal/service"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func isValidDate(date string) bool {
	if date == "" {
		return true
	}
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}

// parseISTAsUTCBoundaries converts a YYYY-MM-DD string (assumed IST) to a UTC time.Time boundary.
func parseISTAsUTCBoundaries(dateStr string, isEnd bool) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}

	// IST is UTC +5:30
	ist := time.FixedZone("IST", 5*3600+1800)

	var istMoment time.Time
	if isEnd {
		istMoment = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, ist)
	} else {
		istMoment = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, ist)
	}

	utcMoment := istMoment.UTC()
	return &utcMoment, nil
}

type AutomationHandler struct {
	templatesService *TemplatesService
	messagesService  *MessagesService
	mappingService   *WebhookMappingService
	orderService     *service.OrderService
	customerService  *service.CustomerService
	settings         *config.SettingsProvider
}

func NewAutomationHandler(tService *TemplatesService, mService *MessagesService, mappingService *WebhookMappingService, orderService *service.OrderService, customerService *service.CustomerService, settings *config.SettingsProvider) *AutomationHandler {
	return &AutomationHandler{
		templatesService: tService,
		messagesService:  mService,
		mappingService:   mappingService,
		orderService:     orderService,
		customerService:  customerService,
		settings:         settings,
	}
}

func (h *AutomationHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Assuming storeID "1" for now
	id, err := h.templatesService.CreateTemplate("1", req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "success": true})
}

func (h *AutomationHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	start, err := parseISTAsUTCBoundaries(startDateStr, false)
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}
	end, err := parseISTAsUTCBoundaries(endDateStr, true)
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	log.Printf("GetTemplates called for storeID: 1, start: %v, end: %v", start, end)
	// DECOUPLED: Removed s.templatesService.SyncStatus("1") to eliminate GET latency.

	templates, err := h.templatesService.GetTemplates("1", start, end)
	if err != nil {
		log.Printf("Error fetching templates: %v", err)
		http.Error(w, "Failed to fetch templates", http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d templates in database", len(templates))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func (h *AutomationHandler) SyncTemplateStatus(w http.ResponseWriter, r *http.Request) {
	log.Printf("SyncTemplateStatus called for storeID: 1")
	err := h.templatesService.SyncStatus("1")
	if err != nil {
		log.Printf("Error syncing statuses: %v", err)
		http.Error(w, "Failed to sync template statuses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Template statuses synced with Meta"})
}

func (h *AutomationHandler) WhatsAppWebhook(w http.ResponseWriter, r *http.Request) {
	// 1. Hub Challenge for verification
	if r.Method == http.MethodGet {
		challenge := r.URL.Query().Get("hub.challenge")
		if challenge != "" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(challenge))
			return
		}
	}

	// 2. Handle status updates
	if r.Method == http.MethodPost {
		// Read body for both validation and unmarshaling
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading WhatsApp webhook body: %v", err)
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// 3. Security: Validate X-Hub-Signature-256
		signature := r.Header.Get("X-Hub-Signature-256")
		if !h.validateWhatsAppSignature(body, signature) {
			log.Printf("Invalid X-Hub-Signature-256 received")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		log.Printf("WhatsApp Webhook Raw Payload: %s", string(body))

		var payload struct {
			Entry []struct {
				Changes []struct {
					Value struct {
						Statuses []struct {
							ID     string `json:"id"`
							Status string `json:"status"`
						} `json:"statuses"`
						Messages []struct {
							From string `json:"from"`
							ID   string `json:"id"`
							Text struct {
								Body string `json:"body"`
							} `json:"text"`
							Type string `json:"type"`
						} `json:"messages"`
						Contacts []struct {
							Profile struct {
								Name string `json:"name"`
							} `json:"profile"`
							WaID string `json:"wa_id"`
						} `json:"contacts"`
					} `json:"value"`
				} `json:"changes"`
			} `json:"entry"`
		}

		if err := json.Unmarshal(body, &payload); err != nil {
			log.Printf("Error unmarshaling WhatsApp webhook: %v", err)
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}


		for _, entry := range payload.Entry {
			for _, change := range entry.Changes {
				// Handle status updates
				for _, status := range change.Value.Statuses {
					err := h.messagesService.HandleStatusUpdate(status.ID, status.Status)
					if err != nil {
						log.Printf("Error updating message status for %s: %v", status.ID, err)
					}
				}

				// Handle incoming messages
				for _, msg := range change.Value.Messages {
					contactName := ""
					for _, contact := range change.Value.Contacts {
						if contact.WaID == msg.From {
							contactName = contact.Profile.Name
							break
						}
					}

					text := ""
					if msg.Type == "text" {
						text = msg.Text.Body
					} else {
						text = fmt.Sprintf("[%s message]", msg.Type)
					}

					valBytes, _ := json.Marshal(change.Value)
					err := h.messagesService.HandleIncomingMessage(msg.From, contactName, msg.ID, text, msg.Type, valBytes)
					if err != nil {
						log.Printf("Error handling incoming message from %s: %v", msg.From, err)
					}
				}
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (h *AutomationHandler) validateWhatsAppSignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	// Signature format: sha256=HEX_DIGEST
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	actualHash := signature[7:]

	mac := hmac.New(sha256.New, []byte(h.settings.GetWhatsAppAppSecret()))
	mac.Write(body)
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(actualHash), []byte(expectedHash))
}

func (h *AutomationHandler) GetTriggers(w http.ResponseWriter, r *http.Request) {
	triggers, err := h.templatesService.GetTriggers("1")
	if err != nil {
		http.Error(w, "Failed to fetch triggers", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(triggers)
}

func (h *AutomationHandler) CreateTrigger(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic      string `json:"webhook_topic"`
		TemplateID int    `json:"template_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	err := h.templatesService.CreateTrigger("1", req.Topic, req.TemplateID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *AutomationHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	start, err := parseISTAsUTCBoundaries(startDateStr, false)
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}
	end, err := parseISTAsUTCBoundaries(endDateStr, true)
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 25
	}
	offset := (page - 1) * limit

	search := r.URL.Query().Get("search")
	templateName := r.URL.Query().Get("template_name")

	messages, err := h.messagesService.GetMessages("1", start, end, search, templateName, limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	totalCount, err := h.messagesService.GetMessagesCount("1", start, end, search, templateName)
	if err != nil {
		log.Printf("Error fetching message count: %v", err)
	}

	activeTemplates, err := h.messagesService.GetActiveTemplateNamesForFilter("1", start, end, search)
	if err != nil {
		log.Printf("Error fetching active templates: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages":         messages,
		"total_count":      totalCount,
		"active_templates": activeTemplates,
	})
}

func (h *AutomationHandler) GetAutomationMetrics(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	start, err := parseISTAsUTCBoundaries(startDateStr, false)
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}
	end, err := parseISTAsUTCBoundaries(endDateStr, true)
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	metrics, err := h.messagesService.GetAutomationMetrics("1", start, end)
	if err != nil {
		http.Error(w, "Failed to fetch metrics", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *AutomationHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	storeID, ok := r.Context().Value("storeID").(string)
	if !ok || storeID == "" {
		storeID = "1"
	}

	var req AutomationTemplate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding update mappings request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Handler: UpdateTemplateMappings request received for ID: %d", req.ID)

	err := h.templatesService.UpdateTemplateMappings(storeID, req.ID, req.VariableMappings)
	if err != nil {
		log.Printf("UpdateTemplateMappings failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AutomationHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	err := h.templatesService.DeleteTemplate(id, "1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AutomationHandler) UpdateTrigger(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID      int  `json:"id"`
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	err := h.templatesService.UpdateTrigger(req.ID, "1", req.Enabled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AutomationHandler) DeleteTrigger(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	err := h.templatesService.DeleteTrigger(id, "1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AutomationHandler) UploadTemplateMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read entire file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	mimeType := header.Header.Get("Content-Type")

	// Call service to forward to Meta
	handle, err := h.templatesService.UploadMediaBytes(fileBytes, mimeType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload to Meta: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"handle": handle,
	})
}
func (h *AutomationHandler) SyncAutomationMetrics(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	start, err := parseISTAsUTCBoundaries(startDateStr, false)
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}
	end, err := parseISTAsUTCBoundaries(endDateStr, true)
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	log.Printf("SyncAutomationMetrics called: start=%v, end=%v", start, end)

	// Since SyncMetricsFromMeta likely takes strings for the Meta API, we might need to format them back to YYYY-MM-DD
	// But the user said "all logic IST". Meta API handles start/end based on account timezone mostly.
	// We'll keep strings for Meta API calls for now but ensure we pass the correct filtered range to our local update.
	metrics, err := h.messagesService.SyncMetricsFromMeta(startDateStr, endDateStr)
	if err != nil {
		log.Printf("Error syncing metrics: %v", err)
		http.Error(w, fmt.Sprintf("Failed to sync metrics: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *AutomationHandler) SendManualMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OrderID    string `json:"order_id"`
		TemplateID int    `json:"template_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.OrderID == "" || req.TemplateID == 0 {
		http.Error(w, "order_id and template_id are required", http.StatusBadRequest)
		return
	}

	orderIDInt, err := strconv.ParseInt(req.OrderID, 10, 64)
	if err != nil {
		log.Printf("Manual Send Error: Invalid order_id format %s: %v", req.OrderID, err)
		http.Error(w, "Invalid order_id format", http.StatusBadRequest)
		return
	}

	// 1. Fetch current order data
	order, err := h.orderService.GetOrderEntity(orderIDInt)
	if err != nil {
		log.Printf("Manual Send Error: Order %s not found: %v", req.OrderID, err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// 2. Execute send
	err = h.mappingService.ExecuteManualSend("1", req.TemplateID, order)
	if err != nil {
		log.Printf("Manual Send Error: Failed to execute send for order %s, template %d: %v", req.OrderID, req.TemplateID, err)
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func (h *AutomationHandler) SendBulkMarketing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CustomerIDs []uint `json:"customer_ids"`
		TemplateID  int    `json:"template_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.CustomerIDs) == 0 || req.TemplateID == 0 {
		http.Error(w, "customer_ids and template_id are required", http.StatusBadRequest)
		return
	}

	// 1. Fetch template and verify name suffix
	template, err := h.templatesService.repo.GetTemplateByID(req.TemplateID)
	if err != nil || template == nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	suffix := h.settings.GetBulkTemplateSuffix()
	cat := strings.ToUpper(strings.TrimSpace(template.Category))
	isMarketing := cat == "MARKETING"
	
	log.Printf("SendBulkMarketing: Validating template '%s' (Category: '%s', Suffix: '%s')", template.TemplateName, cat, suffix)

	if !isMarketing && (suffix == "" || !strings.HasSuffix(strings.ToLower(template.TemplateName), strings.ToLower(suffix))) {
		msg := fmt.Sprintf("Only templates with category 'MARKETING' or ending with '%s' are allowed for bulk selection (This template is '%s')", suffix, cat)
		log.Printf("SendBulkMarketing: Validation FAILED: %s", msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	log.Printf("SendBulkMarketing: Validation PASSED")

	// 2. Fetch customers
	customers, err := h.customerService.GetCustomersByIDs(r.Context(), req.CustomerIDs)
	if err != nil {
		http.Error(w, "Failed to fetch customers", http.StatusInternalServerError)
		return
	}

	// 3. Process sending (async or batch)
	// For now, we'll do it sequentially but we could use a goroutine if it's too many
	successCount := 0
	for _, cust := range customers {
		if cust.PhoneNumber == "" {
			continue
		}

		// Call the mapping service for marketing templates
		err := h.mappingService.ExecuteMarketingSend(cust.SourceID, template, &cust)
		if err == nil {
			successCount++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"sent":    successCount,
		"total":   len(customers),
	})
}
func (h *AutomationHandler) FetchTemplateFromMeta(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Template name is required", http.StatusBadRequest)
		return
	}

	log.Printf("Handler: FetchTemplateFromMeta called for name: %s", name)
	req, err := h.templatesService.FetchRemoteTemplate(name)
	if err != nil {
		log.Printf("FetchRemoteTemplate failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

func (h *AutomationHandler) SyncAllTemplates(w http.ResponseWriter, r *http.Request) {
	storeID, ok := r.Context().Value("storeID").(string)
	if !ok || storeID == "" {
		storeID = "1" // Fallback to primary store ID
	}

	log.Printf("Handler: SyncAllTemplates called for store_id: %s", storeID)
	err := h.templatesService.SyncAllTemplates(storeID)
	if err != nil {
		log.Printf("Handler: SyncAllTemplates service error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully synced all templates from Meta"})
}

func (h *AutomationHandler) SyncSingleTemplate(w http.ResponseWriter, r *http.Request) {
	storeID, ok := r.Context().Value("storeID").(string)
	if !ok || storeID == "" {
		storeID = "1"
	}

	templateName := r.URL.Query().Get("name")
	if templateName == "" {
		http.Error(w, "missing template name parameter", http.StatusBadRequest)
		return
	}

	if err := h.templatesService.SyncSingleTemplate(storeID, templateName); err != nil {
		log.Printf("Handler: SyncSingleTemplate service error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully imported template from Meta"})
}

func (h *AutomationHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	conversations, err := h.messagesService.GetConversations()
	if err != nil {
		http.Error(w, "Failed to fetch conversations", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

func (h *AutomationHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	convIDStr := r.URL.Query().Get("conversation_id")
	convID, _ := strconv.Atoi(convIDStr)
	if convID == 0 {
		http.Error(w, "conversation_id is required", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	messages, err := h.messagesService.GetChatMessages(convID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch chat messages", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (h *AutomationHandler) SendFreeTextMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PhoneNumber string `json:"phone_number"`
		Text        string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PhoneNumber == "" || req.Text == "" {
		http.Error(w, "phone_number and text are required", http.StatusBadRequest)
		return
	}

	// Determine sender role from context/token if available, else default to 'human'
	senderRole := "human"
	
	_, err := h.messagesService.SendFreeTextMessage(req.PhoneNumber, req.Text, senderRole)
	if err != nil {
		log.Printf("SendFreeTextMessage Error: %v", err)
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func (h *AutomationHandler) UpdateConversationMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID   int    `json:"id"`
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ID == 0 || (req.Mode != "auto" && req.Mode != "human") {
		http.Error(w, "valid id and mode ('auto' or 'human') are required", http.StatusBadRequest)
		return
	}

	err := h.messagesService.UpdateConversationMode(req.ID, req.Mode)
	if err != nil {
		http.Error(w, "Failed to update mode", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}
