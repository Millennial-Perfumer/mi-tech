package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	mcpPkg "mi-tech/internal/mcp"
)

// MachineKeyHandler manages machine API keys via admin-only endpoints.
type MachineKeyHandler struct {
	keyService *mcpPkg.MachineKeyService
}

// NewMachineKeyHandler creates a new MachineKeyHandler.
func NewMachineKeyHandler(keyService *mcpPkg.MachineKeyService) *MachineKeyHandler {
	return &MachineKeyHandler{keyService: keyService}
}

// createKeyRequest is the body for POST /api/mcp/keys.
type createKeyRequest struct {
	Name            string   `json:"name"`
	Scopes          []string `json:"scopes"`
	RateLimitPerMin int      `json:"rate_limit_per_min"`
	ExpiresAt       string   `json:"expires_at"` // RFC3339 or empty
}

// ListKeys handles GET /api/mcp/keys.
// @Summary List machine API keys
// @Description List machine API keys (hashes never exposed).
// @Tags mcp
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /mcp/keys [get]
func (h *MachineKeyHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := h.keyService.List()
	if err != nil {
		http.Error(w, "Failed to list machine keys", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "keys": keys})
}

// CreateKey handles POST /api/mcp/keys.
// @Summary Create a machine API key
// @Description Create a machine API key; the plaintext key is returned exactly once.
// @Tags mcp
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{}
// @Router /mcp/keys [post]
func (h *MachineKeyHandler) CreateKey(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if len(req.Scopes) == 0 {
		http.Error(w, "at least one scope is required", http.StatusBadRequest)
		return
	}
	if err := h.keyService.ValidateScopes(req.Scopes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			http.Error(w, "expires_at must be RFC3339", http.StatusBadRequest)
			return
		}
		expiresAt = &t
	}

	plaintext, key, err := h.keyService.Generate(mcpPkg.KeyOptions{
		Name:            req.Name,
		Scopes:          req.Scopes,
		RateLimitPerMin: req.RateLimitPerMin,
		ExpiresAt:       expiresAt,
	})
	if err != nil {
		log.Printf("MachineKeyHandler.CreateKey: %v", err)
		http.Error(w, "Failed to create machine key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"key":        key,
		"plaintext":  plaintext, // shown exactly once
		"key_prefix": mcpPkg.KeyPrefix,
	})
}

// RevokeKey handles DELETE /api/mcp/keys/{id}.
// @Summary Revoke a machine API key
// @Tags mcp
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Router /mcp/keys/{id} [delete]
func (h *MachineKeyHandler) RevokeKey(w http.ResponseWriter, r *http.Request) {
	id, ok := parseKeyID(w, r)
	if !ok {
		return
	}
	if err := h.keyService.Revoke(id); err != nil {
		log.Printf("MachineKeyHandler.RevokeKey: %v", err)
		http.Error(w, "Failed to revoke machine key", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "key revoked"})
}

// RotateKey handles POST /api/mcp/keys/{id}/rotate.
// @Summary Rotate a machine API key
// @Description Replace a key's secret material; the new plaintext is returned once.
// @Tags mcp
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /mcp/keys/{id}/rotate [post]
func (h *MachineKeyHandler) RotateKey(w http.ResponseWriter, r *http.Request) {
	id, ok := parseKeyID(w, r)
	if !ok {
		return
	}
	plaintext, err := h.keyService.Rotate(id)
	if err != nil {
		log.Printf("MachineKeyHandler.RotateKey: %v", err)
		if err == mcpPkg.ErrKeyRevoked {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to rotate machine key", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"plaintext": plaintext,
	})
}

// parseKeyID extracts the trailing id segment from /api/mcp/keys/{id}(/rotate).
func parseKeyID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	trimmed := strings.TrimSuffix(r.URL.Path, "/rotate")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) == 0 {
		http.Error(w, "missing key id", http.StatusBadRequest)
		return 0, false
	}
	idStr := parts[len(parts)-1]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid key id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}
