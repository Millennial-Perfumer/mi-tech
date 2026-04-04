package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"mi-tech/internal/service"
)

// SystemHandler handles system-level API requests.
type SystemHandler struct {
	systemService *service.SystemService
}

// NewSystemHandler creates a new SystemHandler.
func NewSystemHandler(systemService *service.SystemService) *SystemHandler {
	return &SystemHandler{systemService: systemService}
}

// ListDocs returns a list of all documentation files.
// @Summary List documentation slugs
// @Description Retrieve a list of all available documentation files in the /docs directory.
// @Tags discovery
// @Security Bearer
// @Produce json
// @Success 200 {array} service.Document
// @Router /system/docs [get]
func (h *SystemHandler) ListDocs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	docs, err := h.systemService.ListDocs()
	if err != nil {
		http.Error(w, "Failed to list documentation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

// GetDoc handles GET /api/system/docs/{slug}.
// @Summary Get document content
// @Description Fetch the raw Markdown content of a documentation file by its slug.
// @Tags discovery
// @Security Bearer
// @Produce text/markdown
// @Param slug path string true "Document slug (e.g. api-auth)"
// @Success 200 {string} string "Markdown content"
// @Failure 404 {string} string "Document not found"
// @Router /system/docs/{slug} [get]
func (h *SystemHandler) GetDoc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Assuming slug is passed after /api/system/docs/
	slug := strings.TrimPrefix(r.URL.Path, "/api/system/docs/")
	if slug == "" {
		http.Error(w, "Slug is required", http.StatusBadRequest)
		return
	}

	content, err := h.systemService.GetDocContent(slug)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "text/markdown")
	w.Write([]byte(content))
}
