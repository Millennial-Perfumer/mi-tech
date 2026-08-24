package mcp

import (
	"sync"
	"time"

	"mi-tech/internal/mcp/entity"
	"mi-tech/internal/mcp/repository"
)

// AuditEntry carries the non-secret details of a single MCP invocation.
// It deliberately excludes the API key and any authorization header.
type AuditEntry struct {
	KeyID      int64
	KeyName    string
	Scopes     []string
	Tool       string
	Outcome    string
	Status     int
	RemoteIP   string
	RequestID  string
	DurationMs int64
}

// AuditService records MCP invocations asynchronously so tool latency is not
// impacted by database writes. The repository only ever sees hashes/metadata.
type AuditService struct {
	repo      repository.AuditLogRepository
	ch        chan AuditEntry
	stop      chan struct{}
	done      chan struct{}
	closeOnce sync.Once
}

// NewAuditService creates an audit service and starts its worker goroutine.
func NewAuditService(repo repository.AuditLogRepository, buffer int) *AuditService {
	if buffer <= 0 {
		buffer = 256
	}
	s := &AuditService{
		repo: repo,
		ch:   make(chan AuditEntry, buffer),
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go s.run()
	return s
}

// Log enqueues an audit entry for asynchronous persistence.
func (s *AuditService) Log(entry AuditEntry) {
	select {
	case s.ch <- entry:
	default:
		// Drop on overflow rather than block tool execution.
	}
}

// Close stops the worker goroutine, flushing queued entries best-effort.
func (s *AuditService) Close() {
	s.closeOnce.Do(func() { close(s.stop) })
	<-s.done
}

// run drains the audit queue.
func (s *AuditService) run() {
	for {
		select {
		case <-s.stop:
			for {
				select {
				case entry := <-s.ch:
					s.write(entry)
				default:
					close(s.done)
					return
				}
			}
		case entry := <-s.ch:
			s.write(entry)
		}
	}
}

func (s *AuditService) write(entry AuditEntry) {
	scopes := entry.Scopes
	if scopes == nil {
		scopes = []string{}
	}
	log := &entity.MCPAuditLog{
		KeyID:      entry.KeyID,
		KeyName:    entry.KeyName,
		Scopes:     scopes,
		Tool:       entry.Tool,
		Outcome:    entry.Outcome,
		Status:     entry.Status,
		RemoteIP:   entry.RemoteIP,
		RequestID:  entry.RequestID,
		DurationMs: entry.DurationMs,
		CreatedAt:  time.Now(),
	}
	_ = s.repo.Create(log)
}
