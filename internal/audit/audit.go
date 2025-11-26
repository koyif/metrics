package audit

import (
	"sync"

	"github.com/koyif/metrics/internal/models"
)

// AuditObserver defines the interface for audit observers
type AuditObserver interface {
	Notify(event models.AuditEvent) error
}

// Manager manages audit observers and notifies them of events (Subject in Observer pattern)
type Manager struct {
	observers []AuditObserver
	mu        sync.RWMutex
}

// NewManager creates a new audit manager
func NewManager() *Manager {
	return &Manager{
		observers: make([]AuditObserver, 0),
	}
}

// AddObserver adds an observer to the manager
func (m *Manager) AddObserver(observer AuditObserver) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.observers = append(m.observers, observer)
}

// NotifyAll notifies all observers about an audit event
func (m *Manager) NotifyAll(event models.AuditEvent) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, observer := range m.observers {
		go func(obs AuditObserver) {
			_ = obs.Notify(event)
		}(observer)
	}
}

// IsEnabled returns true if there are any observers registered
func (m *Manager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.observers) > 0
}
