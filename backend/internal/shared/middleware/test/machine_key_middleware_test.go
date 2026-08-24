package test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mcpPkg "mi-tech/internal/mcp"
	"mi-tech/internal/mcp/entity"
	"mi-tech/internal/shared/middleware"
)

// memKeyRepo is an in-memory MachineKeyRepository for middleware tests.
type memKeyRepo struct {
	keys []entity.MachineAPIKey
}

func (m *memKeyRepo) Create(key *entity.MachineAPIKey) error {
	key.ID = int64(len(m.keys) + 1)
	m.keys = append(m.keys, *key)
	return nil
}

func (m *memKeyRepo) FindByHash(hash string) (*entity.MachineAPIKey, error) {
	for i := range m.keys {
		if m.keys[i].KeyHash == hash {
			return &m.keys[i], nil
		}
	}
	return nil, errors.New("not found")
}

func (m *memKeyRepo) List() ([]entity.MachineAPIKey, error) { return m.keys, nil }

func (m *memKeyRepo) Update(key *entity.MachineAPIKey) error {
	for i := range m.keys {
		if m.keys[i].ID == key.ID {
			m.keys[i] = *key
			return nil
		}
	}
	return errors.New("not found")
}

func (m *memKeyRepo) TouchLastUsed(id int64, at time.Time) error { return nil }

func newTestKeyService() *mcpPkg.MachineKeyService {
	return mcpPkg.NewMachineKeyService(&memKeyRepo{})
}

func doMachineKeyRequest(svc *mcpPkg.MachineKeyService, h http.Handler, authHeader string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/mcp/tools/list", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func TestMachineKeyMiddlewareRejectsMissing(t *testing.T) {
	svc := newTestKeyService()
	h := middleware.MachineKeyMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := doMachineKeyRequest(svc, h, "")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing header, got %d", rr.Code)
	}
}

func TestMachineKeyMiddlewareRejectsBadFormat(t *testing.T) {
	svc := newTestKeyService()
	h := middleware.MachineKeyMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := doMachineKeyRequest(svc, h, "Basic abc")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for bad format, got %d", rr.Code)
	}
}

func TestMachineKeyMiddlewareRejectsUnknownKey(t *testing.T) {
	svc := newTestKeyService()
	h := middleware.MachineKeyMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := doMachineKeyRequest(svc, h, "Bearer mtk_nope")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unknown key, got %d", rr.Code)
	}
}

func TestMachineKeyMiddlewareAcceptsValidKey(t *testing.T) {
	svc := newTestKeyService()
	plaintext, _, err := svc.Generate(mcpPkg.KeyOptions{Name: "codex", Scopes: []string{mcpPkg.ScopeOrders}})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	h := middleware.MachineKeyMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, _ := r.Context().Value("machineKeyName").(string)
		if name != "codex" {
			t.Errorf("unexpected identity in context: name=%q", name)
		}
		w.WriteHeader(http.StatusOK)
	}))

	rr := doMachineKeyRequest(svc, h, "Bearer "+plaintext)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid key, got %d", rr.Code)
	}
}

func TestMachineKeyMiddlewareRateLimitReturns429(t *testing.T) {
	svc := newTestKeyService()
	plaintext, _, err := svc.Generate(mcpPkg.KeyOptions{
		Name:            "limited",
		Scopes:          []string{mcpPkg.ScopeOrders},
		RateLimitPerMin: 1,
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	h := middleware.MachineKeyMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	if rr := doMachineKeyRequest(svc, h, "Bearer "+plaintext); rr.Code != http.StatusOK {
		t.Fatalf("first request should pass, got %d", rr.Code)
	}
	if rr := doMachineKeyRequest(svc, h, "Bearer "+plaintext); rr.Code != http.StatusTooManyRequests {
		t.Fatalf("second request should return 429, got %d", rr.Code)
	}
}

func TestRequireScopeEnforcement(t *testing.T) {
	svc := newTestKeyService()
	ordersKey, _, err := svc.Generate(mcpPkg.KeyOptions{Name: "codex", Scopes: []string{mcpPkg.ScopeOrders}})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	handler := middleware.MachineKeyMiddleware(svc)(middleware.RequireScope(mcpPkg.ScopeGST)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
	))

	// Key without the gst scope must be rejected with 403.
	rr := doMachineKeyRequest(svc, handler, "Bearer "+ordersKey)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for missing scope, got %d", rr.Code)
	}

	// A key with the required scope must pass.
	gstKey, _, err := svc.Generate(mcpPkg.KeyOptions{Name: "codex2", Scopes: []string{mcpPkg.ScopeGST}})
	if err != nil {
		t.Fatalf("generate gst key: %v", err)
	}
	rr = doMachineKeyRequest(svc, handler, "Bearer "+gstKey)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for authorized scope, got %d", rr.Code)
	}
}

func TestMachineKeyPrefixDistinguishesFromJWT(t *testing.T) {
	// A JWT-shaped token must never be accepted by the machine-key middleware.
	svc := newTestKeyService()
	h := middleware.MachineKeyMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := doMachineKeyRequest(svc, h, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIn0.x")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for JWT token, got %d", rr.Code)
	}
}

// TestKeyPrefixExported ensures the KeyPrefix is a stable public constant.
func TestKeyPrefixExported(t *testing.T) {
	if !strings.HasPrefix("mtk_test", mcpPkg.KeyPrefix) {
		t.Fatal("KeyPrefix must prefix machine keys")
	}
}
