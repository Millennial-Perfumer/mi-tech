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
	settings         *config.SettingsProvider
}

func NewAutomationHandler(tService *TemplatesService, mService *MessagesService, mappingService *WebhookMappingService, orderService *service.OrderService, settings *config.SettingsProvider) *AutomationHandler {
	return &AutomationHandler{
		templatesService: tService,
		messagesService:  mService,
		mappingService:   mappingService,
		orderService:     orderService,
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
				for _, status := range change.Value.Statuses {
					err := h.messagesService.HandleStatusUpdate(status.ID, status.Status)
					if err != nil {
						log.Printf("Error updating message status for %s: %v", status.ID, err)
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

	messages, err := h.messagesService.GetMessages("1", start, end, search, limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	totalCount, err := h.messagesService.GetMessagesCount("1", start, end, search)
	if err != nil {
		log.Printf("Error fetching message count: %v", err)
		// Non-blocking, continue with 0 count or let it fail gracefully
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages":    messages,
		"total_count": totalCount,
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
	var req struct {
		ID int `json:"id"`
		CreateTemplateRequest
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding update request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Handler: UpdateTemplate request received for ID: %d, Name: %s", req.ID, req.Name)

	err := h.templatesService.UpdateTemplate("1", req.ID, req.CreateTemplateRequest)
	if err != nil {
		log.Printf("UpdateTemplate failed: %v", err)
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
