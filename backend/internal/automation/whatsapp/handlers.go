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
	"path/filepath"
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
	agentService     *AgentService
}

func NewAutomationHandler(tService *TemplatesService, mService *MessagesService, mappingService *WebhookMappingService, orderService *service.OrderService, customerService *service.CustomerService, settings *config.SettingsProvider, agentService *AgentService) *AutomationHandler {
	return &AutomationHandler{
		templatesService: tService,
		messagesService:  mService,
		mappingService:   mappingService,
		orderService:     orderService,
		customerService:  customerService,
		settings:         settings,
		agentService:     agentService,
	}
}

// CreateTemplate handles POST /api/automation/whatsapp/templates.
// @Summary Create WhatsApp template
// @Description Import a new template from Meta or create a custom one.
// @Tags automation
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body CreateTemplateRequest true "Template data"
// @Success 200 {object} map[string]interface{}
// @Router /automation/whatsapp/templates [post]
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

	// Assuming storeID config.StoreIDShopify for now
	id, err := h.templatesService.CreateTemplate(config.StoreIDShopify, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "success": true})
}

// GetTemplates handles GET /api/automation/whatsapp/templates.
// @Summary List WhatsApp templates
// @Description Retrieve all imported WhatsApp templates with optional date filtering.
// @Tags automation
// @Security Bearer
// @Produce json
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {array} AutomationTemplate
// @Router /automation/whatsapp/templates [get]
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

	log.Printf("GetTemplates called for storeID: %s, start: %v, end: %v", config.StoreIDShopify, start, end)
	// DECOUPLED: Removed s.templatesService.SyncStatus(config.StoreIDShopify) to eliminate GET latency.

	templates, err := h.templatesService.GetTemplates(config.StoreIDShopify, start, end)
	if err != nil {
		log.Printf("Error fetching templates: %v", err)
		http.Error(w, "Failed to fetch templates", http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d templates in database", len(templates))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

// SyncTemplateStatus handles POST /api/automation/whatsapp/templates/sync.
// @Summary Sync template status
// @Description Fetch latest approval statuses for all templates from Meta.
// @Tags automation
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /automation/whatsapp/templates/sync [post]
func (h *AutomationHandler) SyncTemplateStatus(w http.ResponseWriter, r *http.Request) {
	log.Printf("SyncTemplateStatus called for storeID: %s", config.StoreIDShopify)
	err := h.templatesService.SyncStatus(config.StoreIDShopify)
	if err != nil {
		log.Printf("Error syncing statuses: %v", err)
		http.Error(w, "Failed to sync template statuses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Template statuses synced with Meta"})
}

// WhatsAppWebhook handles Meta Webhooks.
// @Summary WhatsApp Webhook
// @Description Endpoint for Meta Cloud API to push delivery receipts and incoming customer messages.
// @Tags automation
// @Accept json
// @Param X-Hub-Signature-256 header string true "HMAC Signature"
// @Success 200 {string} string "OK"
// @Router /automation/whatsapp/webhook [post]
func (h *AutomationHandler) WhatsAppWebhook(w http.ResponseWriter, r *http.Request) {
	// 1. Hub Challenge for verification
	if r.Method == http.MethodGet {
		query := r.URL.Query()
		mode := query.Get("hub.mode")
		token := query.Get("hub.verify_token")
		challenge := query.Get("hub.challenge")

		expectedToken := h.settings.GetWhatsAppWebhookVerifyToken()

		if mode == "subscribe" && hmac.Equal([]byte(token), []byte(expectedToken)) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(challenge))
			return
		}
		http.Error(w, "Verification failed", http.StatusForbidden)
		return
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
							Type     string `json:"type"`
							Image    *struct { ID string `json:"id"` } `json:"image,omitempty"`
							Video    *struct { ID string `json:"id"` } `json:"video,omitempty"`
							Audio    *struct { ID string `json:"id"` } `json:"audio,omitempty"`
							Document *struct { ID string `json:"id"`; Filename string `json:"filename"` } `json:"document,omitempty"`
							Sticker  *struct { ID string `json:"id"` } `json:"sticker,omitempty"`
							Reaction *struct {
								MessageID string `json:"message_id"`
								Emoji     string `json:"emoji"`
							} `json:"reaction,omitempty"`
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
					var mediaMetadata map[string]interface{}

					switch msg.Type {
					case "text":
						text = msg.Text.Body
					case "image":
						if msg.Image != nil {
							filename, err := h.messagesService.DownloadAndStoreMedia(msg.Image.ID)
							if err == nil {
								text = "Sent an image"
								mediaMetadata = map[string]interface{}{"media_id": msg.Image.ID, "filename": filename}
							} else {
								text = "[Image could not be downloaded]"
							}
						}
					case "video":
						if msg.Video != nil {
							filename, err := h.messagesService.DownloadAndStoreMedia(msg.Video.ID)
							if err == nil {
								text = "Sent a video"
								mediaMetadata = map[string]interface{}{"media_id": msg.Video.ID, "filename": filename}
							} else {
								text = "[Video could not be downloaded]"
							}
						}
					case "audio":
						if msg.Audio != nil {
							filename, err := h.messagesService.DownloadAndStoreMedia(msg.Audio.ID)
							if err == nil {
								text = "Sent an audio message"
								mediaMetadata = map[string]interface{}{"media_id": msg.Audio.ID, "filename": filename}
							} else {
								text = "[Audio could not be downloaded]"
							}
						}
					case "document":
						if msg.Document != nil {
							filename, err := h.messagesService.DownloadAndStoreMedia(msg.Document.ID)
							if err == nil {
								text = fmt.Sprintf("Sent a document: %s", msg.Document.Filename)
								mediaMetadata = map[string]interface{}{
									"media_id": msg.Document.ID, 
									"filename": filename,
									"original_name": msg.Document.Filename,
								}
							} else {
								text = "[Document could not be downloaded]"
							}
						}
					case "sticker":
						if msg.Sticker != nil {
							filename, err := h.messagesService.DownloadAndStoreMedia(msg.Sticker.ID)
							if err == nil {
								text = "Sent a sticker"
								mediaMetadata = map[string]interface{}{"media_id": msg.Sticker.ID, "filename": filename}
							} else {
								text = "[Sticker could not be downloaded]"
							}
						}
					case "reaction":
						if msg.Reaction != nil {
							text = fmt.Sprintf("Reacted with %s", msg.Reaction.Emoji)
							mediaMetadata = map[string]interface{}{
								"reaction_emoji": msg.Reaction.Emoji,
								"reacting_to":    msg.Reaction.MessageID,
							}
						}
					default:
						text = fmt.Sprintf("[%s message]", msg.Type)
					}

					var valBytes []byte
					if mediaMetadata != nil {
						// Merge captured metadata into the raw payload for full context
						base := make(map[string]interface{})
						json.Unmarshal(body, &base)
						base["extracted_metadata"] = mediaMetadata
						valBytes, _ = json.Marshal(base)
					} else {
						valBytes, _ = json.Marshal(change.Value)
					}
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

// TelegramWebhook handles incoming messages from the Telegram bot (Admin interaction).
func (h *AutomationHandler) TelegramWebhook(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Message struct {
			Text string `json:"text"`
			Chat struct {
				ID int64 `json:"id"`
			} `json:"chat"`
		} `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Security: Verify Chat ID
	adminChatIDStr := h.settings.Get("telegram_chat_id")
	adminChatID, _ := strconv.ParseInt(adminChatIDStr, 10, 64)

	if payload.Message.Chat.ID != adminChatID {
		log.Printf("Unauthorized Telegram interaction from ChatID: %d", payload.Message.Chat.ID)
		w.WriteHeader(http.StatusOK) // Silent ignore
		return
	}

	text := strings.TrimSpace(payload.Message.Text)
	if strings.HasPrefix(text, "/concerns") {
		summary, err := h.agentService.GenerateDailyConcernsSummary()
		if err != nil {
			h.agentService.notifService.SendSummary(fmt.Sprintf("❌ Error generating report: %v", err))
		} else {
			h.agentService.notifService.SendSummary(summary)
		}
	}

	w.WriteHeader(http.StatusOK)
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

// GetTriggers handles GET /api/automation/whatsapp/triggers.
// @Summary List automation triggers
// @Description Retrieve all configured Webhook-to-Template mappings.
// @Tags automation
// @Security Bearer
// @Produce json
// @Success 200 {array} Trigger
// @Router /automation/whatsapp/triggers [get]
func (h *AutomationHandler) GetTriggers(w http.ResponseWriter, r *http.Request) {
	triggers, err := h.templatesService.GetTriggers(config.StoreIDShopify)
	if err != nil {
		http.Error(w, "Failed to fetch triggers", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(triggers)
}

// CreateTrigger handles POST /api/automation/whatsapp/triggers.
// @Summary Create automation trigger
// @Description Link a Shopify webhook topic to a WhatsApp template.
// @Tags automation
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body object true "Trigger data"
// @Success 201 "Created"
// @Router /automation/whatsapp/triggers [post]
func (h *AutomationHandler) CreateTrigger(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic      string `json:"webhook_topic"`
		TemplateID int    `json:"template_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	err := h.templatesService.CreateTrigger(config.StoreIDShopify, req.Topic, req.TemplateID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// GetMessages handles GET /api/automation/whatsapp/messages.
// @Summary List message logs
// @Description Retrieve a paginated list of all sent and received WhatsApp messages.
// @Tags automation
// @Security Bearer
// @Produce json
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search term"
// @Param template_name query string false "Filter by template"
// @Success 200 {object} map[string]interface{}
// @Router /automation/whatsapp/messages [get]
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

	messages, err := h.messagesService.GetMessages(config.StoreIDShopify, start, end, search, templateName, limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	totalCount, err := h.messagesService.GetMessagesCount(config.StoreIDShopify, start, end, search, templateName)
	if err != nil {
		log.Printf("Error fetching message count: %v", err)
	}

	activeTemplates, err := h.messagesService.GetActiveTemplateNamesForFilter(config.StoreIDShopify, start, end, search)
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

// GetAutomationMetrics handles GET /api/automation/whatsapp/metrics.
// @Summary Get automation analytics
// @Description Aggregate delivery, read, and failure rates for automated messages.
// @Tags automation
// @Security Bearer
// @Produce json
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} map[string]interface{}
// @Router /automation/whatsapp/metrics [get]
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

	metrics, err := h.messagesService.GetAutomationMetrics(config.StoreIDShopify, start, end)
	if err != nil {
		http.Error(w, "Failed to fetch metrics", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// UpdateTemplate handles PUT /api/automation/whatsapp/templates.
// @Summary Update template mapping
// @Description Update the variable mappings for a WhatsApp template.
// @Tags automation
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body AutomationTemplate true "Updated mapping"
// @Success 200 "OK"
// @Router /automation/whatsapp/templates [put]
func (h *AutomationHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	storeID, ok := r.Context().Value("storeID").(string)
	if !ok || storeID == "" {
		storeID = config.StoreIDShopify
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

// DeleteTemplate handles DELETE /api/automation/whatsapp/templates.
// @Summary Delete WhatsApp template
// @Description Remove a template from the local database.
// @Tags automation
// @Security Bearer
// @Param id query int true "Template ID"
// @Success 200 "OK"
// @Router /automation/whatsapp/templates [delete]
func (h *AutomationHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	err := h.templatesService.DeleteTemplate(id, config.StoreIDShopify)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UpdateTrigger handles PUT /api/automation/whatsapp/triggers.
// @Summary Toggle automation trigger
// @Description Enable or disable a specific Webhook-to-Template trigger.
// @Tags automation
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body object true "Trigger toggle data"
// @Success 200 "OK"
// @Router /automation/whatsapp/triggers [put]
func (h *AutomationHandler) UpdateTrigger(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID      int  `json:"id"`
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	err := h.templatesService.UpdateTrigger(req.ID, config.StoreIDShopify, req.Enabled)
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
	err := h.templatesService.DeleteTrigger(id, config.StoreIDShopify)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UploadTemplateMedia handles POST /api/automation/whatsapp/templates/upload.
// @Summary Upload media to Meta
// @Description Upload an image or PDF to Meta's servers for use in template headers.
// @Tags automation
// @Security Bearer
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Media file"
// @Success 200 {object} map[string]string
// @Router /automation/whatsapp/templates/upload [post]
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

	// ALSO save locally for consistency and persistence (subject to 15-day cleanup)
	_, err = h.messagesService.StoreMedia(handle, fileBytes, mimeType)
	if err != nil {
		log.Printf("Warning: Failed to save template media locally: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"handle": handle,
	})
}
// SyncAutomationMetrics handles POST /api/automation/whatsapp/metrics/sync.
// @Summary Sync Meta insights
// @Description Fetch delivery and read analytics directly from Meta's Insight API.
// @Tags automation
// @Security Bearer
// @Produce json
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} map[string]interface{}
// @Router /automation/whatsapp/metrics/sync [post]
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

// SendManualMessage handles POST /api/automation/whatsapp/send-manual.
// @Summary Send manual template
// @Description Trigger a specific template message for a given order manually.
// @Tags automation
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body object true "Send request (order_id, template_id)"
// @Success 200 {object} map[string]interface{}
// @Router /automation/whatsapp/send-manual [post]
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
	err = h.mappingService.ExecuteManualSend(config.StoreIDShopify, req.TemplateID, order)
	if err != nil {
		log.Printf("Manual Send Error: Failed to execute send for order %s, template %d: %v", req.OrderID, req.TemplateID, err)
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// SendBulkMarketing handles POST /api/automation/whatsapp/send-bulk.
// @Summary Send bulk marketing
// @Description Dispatch a template message to multiple customers simultaneously.
// @Tags automation
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body object true "Bulk send request"
// @Success 200 {object} map[string]interface{}
// @Router /automation/whatsapp/send-bulk [post]
// UploadChatMedia handles POST /api/automation/whatsapp/chat/upload.
func (h *AutomationHandler) UploadChatMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(20 << 20) // 20 MB limit
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

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	mimeType := header.Header.Get("Content-Type")
	filename := header.Filename

	// 1. Forward to Meta
	id, err := h.messagesService.metaClient.UploadWhatsAppMedia(fileBytes, filename, mimeType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload to Meta: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. ALSO save locally for better UI rendering
	localFilename, err := h.messagesService.StoreMedia(id, fileBytes, mimeType)
	if err != nil {
		log.Printf("Warning: Failed to save outgoing media locally: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"media_id": id,
		"filename": localFilename,
	})
}

// SendChatMedia handles POST /api/automation/whatsapp/chat/send-media.
func (h *AutomationHandler) SendChatMedia(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PhoneNumber string `json:"phone_number"`
		MediaID     string `json:"media_id"`
		Type        string `json:"type"`
		Caption     string `json:"caption"`
		Filename    string `json:"filename"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Create metadata for the outgoing message to include the local filename
	metadata := map[string]interface{}{
		"media_id": req.MediaID,
		"caption":  req.Caption,
		"filename": req.Filename,
	}
	metadataBytes, _ := json.Marshal(metadata)

	displayText := fmt.Sprintf("Sent a %s", req.Type)
	if req.Caption != "" {
		displayText = req.Caption
	}

	// 1. Send via Meta
	msgID, err := h.messagesService.metaClient.SendMediaMessage(req.PhoneNumber, req.MediaID, req.Type, req.Caption)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send media via Meta: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. Upsert conversation
	convID, err := h.messagesService.repo.UpsertConversation(req.PhoneNumber, "", displayText)
	if err != nil {
		http.Error(w, "Failed to update conversation", http.StatusInternalServerError)
		return
	}

	// 3. Save message with metadata
	chatMsg := ChatMessage{
		ConversationID: convID,
		MessageID:      msgID,
		Text:           displayText,
		Type:           req.Type,
		Direction:      "outgoing",
		SenderRole:     "human",
		Status:         "sent",
		SentAt:         time.Now().UTC(),
		Metadata:       metadataBytes,
	}
	_, err = h.messagesService.repo.SaveChatMessage(chatMsg)
	if err != nil {
		http.Error(w, "Failed to save outgoing media message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
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

// SyncAllTemplates handles POST /api/automation/whatsapp/templates/sync-all.
// @Summary Full template import
// @Description Wipe and re-import all templates from the linked Meta WABA.
// @Tags automation
// @Security Bearer
// @Produce json
// @Success 200 "OK"
// @Router /automation/whatsapp/templates/sync-all [post]
func (h *AutomationHandler) SyncAllTemplates(w http.ResponseWriter, r *http.Request) {
	storeID, ok := r.Context().Value("storeID").(string)
	if !ok || storeID == "" {
		storeID = config.StoreIDShopify // Fallback to primary store ID
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
		storeID = config.StoreIDShopify
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

// GetConversations handles GET /api/automation/whatsapp/conversations.
// @Summary List chat threads
// @Description Retrieve all active customer chat conversations.
// @Tags automation
// @Security Bearer
// @Produce json
// @Success 200 {array} Conversation
// @Router /automation/whatsapp/conversations [get]
func (h *AutomationHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	conversations, err := h.messagesService.GetConversations()
	if err != nil {
		http.Error(w, "Failed to fetch conversations", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

// GetChatMessages handles GET /api/automation/whatsapp/chat.
// @Summary Get chat history
// @Description Retrieve all messages for a specific conversation.
// @Tags automation
// @Security Bearer
// @Produce json
// @Param conversation_id query int true "Conv ID"
// @Success 200 {array} ChatMessage
// @Router /automation/whatsapp/chat [get]
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

// SendFreeTextMessage handles POST /api/automation/whatsapp/send-message.
// @Summary Send free-text
// @Description Send a non-template, free-text message to a customer (Human-to-Human).
// @Tags automation
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body object true "Message data"
// @Success 200 "OK"
// @Router /automation/whatsapp/send-message [post]
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

// UpdateConversationMode handles PUT /api/automation/whatsapp/conversations/mode.
// @Summary Toggle chat mode
// @Description Switch between 'auto' (bot) and 'human' (agent) mode for a chat thread.
// @Tags automation
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body object true "Mode data"
// @Success 200 "OK"
// @Router /automation/whatsapp/conversations/mode [put]
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

// GetWhatsAppMedia serves local WhatsApp media files.
// GetWhatsAppMedia serves local WhatsApp media files.
func (h *AutomationHandler) GetWhatsAppMedia(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Basic security check to prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		http.Error(w, "Invalid filename", http.StatusForbidden)
		return
	}

	dir := filepath.Join("uploads", "whatsapp")
	path := filepath.Join(dir, filename)

	// Verify the file exists within the dedicated uploads directory
	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusInternalServerError)
		return
	}
	
	absDir, _ := filepath.Abs(dir)
	if !strings.HasPrefix(absPath, absDir) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, path)
}

func (h *AutomationHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.templatesService.GetEvents()
	if err != nil {
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (h *AutomationHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req AutomationEvent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: Failed to decode CreateEvent request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Topic == "" {
		log.Printf("ERROR: CreateEvent missing required fields: name='%s', topic='%s'", req.Name, req.Topic)
		http.Error(w, "name and topic are required", http.StatusBadRequest)
		return
	}

	log.Printf("DEBUG: Attempting to save event: %+v", req)
	err := h.templatesService.SaveEvent(req)
	if err != nil {
		log.Printf("ERROR: Failed to save event to database: %v", err)
		http.Error(w, fmt.Sprintf("Failed to save event: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("SUCCESS: Event created/updated: %s (%s)", req.Name, req.Topic)
	w.WriteHeader(http.StatusCreated)
}

func (h *AutomationHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	if err := h.templatesService.DeleteEvent(id); err != nil {
		http.Error(w, "Failed to delete event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetOrderMessages handles GET /api/automation/whatsapp/messages/order.
// @Summary List messages for an order
// @Description Retrieve all automation messages sent for a specific order.
// @Tags automation
// @Security Bearer
// @Produce json
// @Param order_id query string true "Order ID"
// @Success 200 {array} AutomationMessage
// @Router /automation/whatsapp/messages/order [get]
func (h *AutomationHandler) GetOrderMessages(w http.ResponseWriter, r *http.Request) {
	orderIDStr := r.URL.Query().Get("order_id")
	if orderIDStr == "" {
		http.Error(w, "order_id is required", http.StatusBadRequest)
		return
	}

	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		// Try flexible lookup if it's an external ID
		order, err := h.orderService.GetOrderFlexible(orderIDStr)
		if err != nil {
			http.Error(w, "Invalid order_id format", http.StatusBadRequest)
			return
		}
		orderID = order.ID
	}

	messages, err := h.messagesService.GetMessagesByOrderID(orderID)
	if err != nil {
		http.Error(w, "Failed to fetch order messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

