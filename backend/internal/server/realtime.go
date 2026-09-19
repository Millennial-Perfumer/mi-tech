package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type realtimeEvent struct {
	Type string `json:"type"`
	At   string `json:"at"`
}

type realtimeClient struct {
	conn *websocket.Conn
	send chan realtimeEvent
}

type realtimeHub struct {
	mu      sync.Mutex
	clients map[*realtimeClient]struct{}
	tickets map[string]time.Time
}

func newRealtimeHub() *realtimeHub {
	return &realtimeHub{clients: make(map[*realtimeClient]struct{}), tickets: make(map[string]time.Time)}
}

func (h *realtimeHub) issueTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		http.Error(w, "could not create socket ticket", http.StatusInternalServerError)
		return
	}
	ticket := base64.RawURLEncoding.EncodeToString(raw)
	h.mu.Lock()
	now := time.Now()
	for key, expires := range h.tickets {
		if !expires.After(now) {
			delete(h.tickets, key)
		}
	}
	h.tickets[ticket] = now.Add(30 * time.Second)
	h.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(map[string]string{"ticket": ticket})
}

func (h *realtimeHub) consumeTicket(ticket string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	expires, ok := h.tickets[ticket]
	delete(h.tickets, ticket)
	return ok && expires.After(time.Now())
}

func realtimeOriginAllowed(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	parsed, err := url.Parse(origin)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return false
	}
	if parsed.Host == r.Host {
		return true
	}
	// Vite's /api proxy keeps the browser origin on its own loopback port.
	if (parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1") &&
		(strings.HasPrefix(r.Host, "localhost:") || strings.HasPrefix(r.Host, "127.0.0.1:")) {
		return true
	}
	for _, allowed := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		if strings.TrimRight(strings.TrimSpace(allowed), "/") == origin {
			return true
		}
	}
	return false
}

func (h *realtimeHub) serveWS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !realtimeOriginAllowed(r) {
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return
	}
	if !h.consumeTicket(r.URL.Query().Get("ticket")) {
		http.Error(w, "invalid or expired socket ticket", http.StatusUnauthorized)
		return
	}
	// The API server uses short HTTP timeouts. Clear them before the upgrade;
	// the WebSocket read loop below installs its own bounded deadlines.
	controller := http.NewResponseController(w)
	_ = controller.SetReadDeadline(time.Time{})
	_ = controller.SetWriteDeadline(time.Time{})
	conn, err := (&websocket.Upgrader{CheckOrigin: realtimeOriginAllowed}).Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &realtimeClient{conn: conn, send: make(chan realtimeEvent, 32)}
	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.clients, client)
		h.mu.Unlock()
		_ = conn.Close()
	}()

	// A single writer owns each connection. Client messages are ignored; this
	// channel only tells browsers when to fetch fresh server state.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case event := <-client.send:
				_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteJSON(event); err != nil {
					_ = conn.Close()
					return
				}
			case <-r.Context().Done():
				_ = conn.Close()
				return
			}
		}
	}()
	conn.SetReadLimit(1024)
	_ = conn.SetReadDeadline(time.Now().Add(75 * time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(75 * time.Second))
	})
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				_ = conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second))
			case <-done:
				return
			}
		}
	}()
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *realtimeHub) publish(kind string) {
	event := realtimeEvent{Type: kind, At: time.Now().UTC().Format(time.RFC3339Nano)}
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		select {
		case client.send <- event:
		default:
			_ = client.conn.Close()
			delete(h.clients, client)
		}
	}
}

type realtimeResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *realtimeResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *realtimeResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(data)
}

func realtimeEventType(path string) string {
	switch {
	case strings.HasPrefix(path, "/api/orders"), strings.HasPrefix(path, "/api/shopify/sync"), strings.HasPrefix(path, "/api/webhooks/shopify"):
		return "orders.changed"
	case strings.HasPrefix(path, "/api/customers"):
		return "customers.changed"
	case strings.HasPrefix(path, "/api/inventory"):
		return "inventory.changed"
	case strings.HasPrefix(path, "/api/automation/whatsapp"), strings.HasPrefix(path, "/api/webhooks"):
		return "communication.changed"
	case strings.HasPrefix(path, "/api/b2b"):
		return "b2b.changed"
	case strings.HasPrefix(path, "/api/support"):
		return "support.changed"
	case strings.HasPrefix(path, "/api/feedback"):
		return "feedback.changed"
	case strings.HasPrefix(path, "/api/abandoned"):
		return "abandoned.changed"
	case strings.HasPrefix(path, "/api/smm-queue"):
		return "smm.changed"
	case strings.HasPrefix(path, "/api/settings"), strings.HasPrefix(path, "/api/configs"):
		return "settings.changed"
	case strings.HasPrefix(path, "/api/users"):
		return "users.changed"
	case strings.HasPrefix(path, "/api/marketing"):
		return "marketing.changed"
	default:
		return ""
	}
}

func (h *realtimeHub) notifyWrites(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodOptions || strings.HasPrefix(r.URL.Path, "/mcp") || strings.HasPrefix(r.URL.Path, "/api/realtime/") {
			next.ServeHTTP(w, r)
			return
		}
		kind := realtimeEventType(r.URL.Path)
		if kind == "" {
			next.ServeHTTP(w, r)
			return
		}
		recorder := &realtimeResponseWriter{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		if status >= 200 && status < 300 {
			h.publish(kind)
		}
	})
}
