package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestRealtimeTicketIsSingleUseAndOriginRestricted(t *testing.T) {
	hub := newRealtimeHub()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/realtime/ticket", hub.issueTicket)
	mux.HandleFunc("/api/realtime/ws", hub.serveWS)
	server := httptest.NewServer(mux)
	defer server.Close()

	response, err := http.Post(server.URL+"/api/realtime/ticket", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var issued struct{ Ticket string }
	if err := json.NewDecoder(response.Body).Decode(&issued); err != nil || issued.Ticket == "" {
		t.Fatalf("ticket response: %v, %+v", err, issued)
	}
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/realtime/ws?ticket=" + url.QueryEscape(issued.Ticket)
	deniedHeader := http.Header{"Origin": []string{"https://not-allowed.example"}}
	if conn, resp, err := websocket.DefaultDialer.Dial(wsURL, deniedHeader); err == nil {
		_ = conn.Close()
		t.Fatal("unexpected socket with forbidden origin")
	} else if resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected forbidden origin, got %v: %v", resp, err)
	}

	header := http.Header{"Origin": []string{server.URL}}
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("valid ticket rejected: %v (%v)", err, resp)
	}
	defer conn.Close()
	if second, resp, err := websocket.DefaultDialer.Dial(wsURL, header); err == nil {
		_ = second.Close()
		t.Fatal("ticket was accepted twice")
	} else if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected single-use rejection, got %v: %v", resp, err)
	}

	hub.publish("orders.changed")
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var event realtimeEvent
	if err := conn.ReadJSON(&event); err != nil || event.Type != "orders.changed" || event.At == "" {
		t.Fatalf("event delivery failed: %v, %+v", err, event)
	}
}

func TestRealtimePublishesOnlySuccessfulWrites(t *testing.T) {
	hub := newRealtimeHub()
	client := &realtimeClient{send: make(chan realtimeEvent, 2)}
	hub.clients[client] = struct{}{}
	handler := hub.notifyWrites(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fail") == "1" {
			http.Error(w, "failed", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	for _, path := range []string{"/api/customers/1?fail=1", "/api/customers/1"} {
		req := httptest.NewRequest(http.MethodPut, path, nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}
	select {
	case event := <-client.send:
		if event.Type != "customers.changed" {
			t.Fatalf("unexpected event: %s", event.Type)
		}
	default:
		t.Fatal("successful write did not publish")
	}
	select {
	case event := <-client.send:
		t.Fatalf("failed write published: %s", event.Type)
	default:
	}
}
