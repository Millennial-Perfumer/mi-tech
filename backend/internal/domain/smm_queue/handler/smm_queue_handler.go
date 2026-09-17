package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"mi-tech/internal/domain/smm_queue/service"
)

const maxMultipartBodyBytes = 55 * 1024 * 1024

type SMMQueueHandler struct {
	service *service.SMMQueueService
}

func NewSMMQueueHandler(queueService *service.SMMQueueService) *SMMQueueHandler {
	return &SMMQueueHandler{service: queueService}
}

// Handle dispatches the collection-level queue operations.
func (h *SMMQueueHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateQueue(w, r)
	case http.MethodGet:
		h.ListQueue(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleItem dispatches detail, media, update, and delete operations for one
// timestamped queue folder.
func (h *SMMQueueHandler) HandleItem(w http.ResponseWriter, r *http.Request) {
	relativePath := strings.TrimPrefix(r.URL.Path, "/api/smm-queue/")
	parts := strings.Split(strings.Trim(relativePath, "/"), "/")
	if len(parts) == 1 && parts[0] != "" {
		folder, err := url.PathUnescape(parts[0])
		if err != nil {
			http.Error(w, "Invalid queue folder", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.GetQueue(w, r, folder)
		case http.MethodPut:
			h.UpdateQueue(w, r, folder)
		case http.MethodDelete:
			h.DeleteQueue(w, r, folder)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	if len(parts) == 3 && parts[1] == "media" && r.Method == http.MethodGet {
		folder, folderErr := url.PathUnescape(parts[0])
		filename, filenameErr := url.PathUnescape(parts[2])
		if folderErr != nil || filenameErr != nil {
			http.Error(w, "Invalid queue media path", http.StatusBadRequest)
			return
		}
		h.GetMedia(w, r, folder, filename)
		return
	}
	http.NotFound(w, r)
}

// CreateQueue accepts caption, hashtags, and repeated media image fields.
// Each successful request creates a timestamp-prefixed virtual folder in Blob
// Storage and publishes _READY.json last for n8n to discover.
func (h *SMMQueueHandler) CreateQueue(w http.ResponseWriter, r *http.Request) {
	caption, hashtags, media, ok := h.parseQueueForm(w, r)
	if !ok {
		return
	}
	result, err := h.service.Queue(r.Context(), caption, hashtags, media)
	if err != nil {
		h.writeServiceError(w, "CreateQueue", err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "queue": result})
}

func (h *SMMQueueHandler) ListQueue(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		http.Error(w, "SMM queue storage is not configured", http.StatusServiceUnavailable)
		return
	}
	items, err := h.service.List(r.Context())
	if err != nil {
		h.writeServiceError(w, "ListQueue", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "items": items})
}

func (h *SMMQueueHandler) GetQueue(w http.ResponseWriter, r *http.Request, folder string) {
	if h.service == nil {
		http.Error(w, "SMM queue storage is not configured", http.StatusServiceUnavailable)
		return
	}
	manifest, err := h.service.Get(r.Context(), folder)
	if err != nil {
		h.writeServiceError(w, "GetQueue", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "status": "queued", "queue": manifest})
}

func (h *SMMQueueHandler) GetMedia(w http.ResponseWriter, r *http.Request, folder, filename string) {
	if h.service == nil {
		http.Error(w, "SMM queue storage is not configured", http.StatusServiceUnavailable)
		return
	}
	data, contentType, err := h.service.GetMedia(r.Context(), folder, filename)
	if err != nil {
		h.writeServiceError(w, "GetMedia", err)
		return
	}
	if contentType == "" && len(data) > 0 {
		contentType = http.DetectContentType(data[:min(len(data), 512)])
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "inline; filename=\""+filename+"\"")
	_, _ = w.Write(data)
}

func (h *SMMQueueHandler) UpdateQueue(w http.ResponseWriter, r *http.Request, folder string) {
	caption, hashtags, media, ok := h.parseQueueForm(w, r)
	if !ok {
		return
	}
	result, err := h.service.Update(r.Context(), folder, caption, hashtags, media)
	if err != nil {
		h.writeServiceError(w, "UpdateQueue", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "queue": result})
}

func (h *SMMQueueHandler) DeleteQueue(w http.ResponseWriter, r *http.Request, folder string) {
	if h.service == nil {
		http.Error(w, "SMM queue storage is not configured", http.StatusServiceUnavailable)
		return
	}
	if err := h.service.Delete(r.Context(), folder); err != nil {
		h.writeServiceError(w, "DeleteQueue", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SMMQueueHandler) parseQueueForm(w http.ResponseWriter, r *http.Request) (string, string, []service.MediaUpload, bool) {
	if h.service == nil {
		http.Error(w, "SMM queue storage is not configured", http.StatusServiceUnavailable)
		return "", "", nil, false
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartBodyBytes)
	if err := r.ParseMultipartForm(maxMultipartBodyBytes); err != nil {
		if errors.Is(err, http.ErrNotMultipart) {
			http.Error(w, "multipart/form-data is required", http.StatusBadRequest)
			return "", "", nil, false
		}
		http.Error(w, "The upload is too large or invalid", http.StatusBadRequest)
		return "", "", nil, false
	}

	files := r.MultipartForm.File["media"]
	if len(files) == 0 {
		http.Error(w, "At least one image is required", http.StatusBadRequest)
		return "", "", nil, false
	}
	if len(files) > service.MaxMediaFiles {
		http.Error(w, "Too many images", http.StatusBadRequest)
		return "", "", nil, false
	}

	media := make([]service.MediaUpload, 0, len(files))
	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			http.Error(w, "Unable to read an uploaded image", http.StatusBadRequest)
			return "", "", nil, false
		}
		data, readErr := io.ReadAll(io.LimitReader(file, service.MaxFileBytes+1))
		closeErr := file.Close()
		if readErr != nil || closeErr != nil {
			http.Error(w, "Unable to read an uploaded image", http.StatusBadRequest)
			return "", "", nil, false
		}
		media = append(media, service.MediaUpload{
			Filename:    header.Filename,
			ContentType: header.Header.Get("Content-Type"),
			Data:        data,
		})
	}
	return r.FormValue("caption"), r.FormValue("hashtags"), media, true
}

func (h *SMMQueueHandler) writeServiceError(w http.ResponseWriter, operation string, err error) {
	switch {
	case errors.Is(err, service.ErrStorageUnavailable):
		http.Error(w, "SMM queue storage is not configured", http.StatusServiceUnavailable)
	case errors.Is(err, service.ErrInvalidQueue):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, service.ErrQueueNotFound):
		http.Error(w, "SMM queue item not found", http.StatusNotFound)
	default:
		log.Printf("SMMQueueHandler.%s error: %v", operation, err)
		http.Error(w, "Unable to access the SMM queue", http.StatusBadGateway)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("SMM queue response error: %v", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
