package whatsapp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type AutomationHandler struct {
	templatesService *TemplatesService
	messagesService  *MessagesService
}

func NewAutomationHandler(tService *TemplatesService, mService *MessagesService) *AutomationHandler {
	return &AutomationHandler{
		templatesService: tService,
		messagesService:  mService,
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
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	log.Printf("GetTemplates called for storeID: 1, start: %s, end: %s", startDate, endDate)
	// Sync status before returning
	h.templatesService.SyncStatus("1")

	templates, err := h.templatesService.GetTemplates("1", startDate, endDate)
	if err != nil {
		log.Printf("Error fetching templates: %v", err)
		http.Error(w, "Failed to fetch templates", http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d templates in database", len(templates))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func (h *AutomationHandler) WhatsAppWebhook(w http.ResponseWriter, r *http.Request) {
	// 1. Hub Challenge for verification (Optional, but good for setup)
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
		body, _ := io.ReadAll(r.Body)
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
					h.messagesService.HandleStatusUpdate(status.ID, status.Status)
				}
			}
		}

		w.WriteHeader(http.StatusOK)
	}
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
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	messages, err := h.messagesService.GetMessages("1", startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (h *AutomationHandler) GetAutomationMetrics(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	metrics, err := h.messagesService.GetAutomationMetrics("1", startDate, endDate)
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
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	log.Printf("SyncAutomationMetrics called: start=%s, end=%s", startDate, endDate)

	metrics, err := h.messagesService.SyncMetricsFromMeta(startDate, endDate)
	if err != nil {
		log.Printf("Error syncing metrics: %v", err)
		http.Error(w, fmt.Sprintf("Failed to sync metrics: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
