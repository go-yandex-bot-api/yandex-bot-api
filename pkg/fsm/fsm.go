// Package fsm provides finite state machine implementations.
package fsm

import (
	"sync"
	"time"
)

// Storage defines the interface for FSM (Finite State Machine) storage.
// This allows developers to use Memory, Redis, PostgreSQL, etc.
type Storage interface {
	Set(userID, state string)
	Get(userID string) string
	Delete(userID string)
	SetData(userID, key string, value any)
	GetData(userID, key string) (any, bool)
}

type stateItem struct {
	state     string
	expiresAt time.Time
	data      map[string]any
}

// MemoryStorage is a thread-safe in-memory storage for FSM with TTL.
type MemoryStorage struct {
	mu       sync.RWMutex
	states   map[string]stateItem
	ttl      time.Duration
	stop     chan struct{} // Канал для остановки фоновой горутины
	stopOnce sync.Once
}

// NewMemoryStorage creates a new ready-to-use MemoryStorage with a specified Time-To-Live.
// If a user abandons the bot, their state will automatically expire after this TTL.
// It starts a background goroutine to clean up expired states automatically.
func NewMemoryStorage(ttl time.Duration) *MemoryStorage {
	m := &MemoryStorage{
		states: make(map[string]stateItem),
		ttl:    ttl,
		stop:   make(chan struct{}),
	}

	if ttl > 0 {
		go m.cleanupLoop() // Запускаем фоновый сборщик мусора
	}

	return m
}

// Stop halts the background garbage collector.
// You should call this if you are destroying the MemoryStorage (usually on bot shutdown).
func (m *MemoryStorage) Stop() {
	m.stopOnce.Do(func() {
		close(m.stop)
	})
}

// cleanupLoop runs in the background and periodically deletes expired states.
func (m *MemoryStorage) cleanupLoop() {
	interval := m.ttl / 2 //nolint:mnd // cleanup interval is half the ttl
	if interval <= 0 {
		interval = time.Second // Prevent panic on NewTicker(0)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stop:
			return // Завершаем горутину, если вызван Stop()
		}
	}
}

// cleanup iterates through the map and deletes expired items.
func (m *MemoryStorage) cleanup() {
	var expired []string
	now := time.Now()

	m.mu.RLock()
	for userID, item := range m.states {
		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			expired = append(expired, userID)
		}
	}
	m.mu.RUnlock()

	if len(expired) == 0 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	for _, userID := range expired {
		if item, exists := m.states[userID]; exists && now.After(item.expiresAt) {
			delete(m.states, userID)
		}
	}
}

// Set saves the state for a specific user and updates their expiration time.
func (m *MemoryStorage) Set(userID, state string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var data map[string]any
	now := time.Now()
	if oldItem, exists := m.states[userID]; exists {
		if m.ttl == 0 || now.Before(oldItem.expiresAt) || now.Equal(oldItem.expiresAt) {
			data = oldItem.data
		}
	}

	item := stateItem{state: state, data: data}
	if m.ttl > 0 {
		item.expiresAt = now.Add(m.ttl)
	}
	m.states[userID] = item
}

// Get retrieves the current state for a specific user.
// Returns an empty string if no state is found or if the state has expired.
func (m *MemoryStorage) Get(userID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.states[userID]
	if !exists {
		return ""
	}

	if m.ttl > 0 && time.Now().After(item.expiresAt) {
		return ""
	}

	return item.state
}

// Delete removes the state for a specific user.
func (m *MemoryStorage) Delete(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.states, userID)
}

// SetData saves an arbitrary key-value pair for a specific user.
func (m *MemoryStorage) SetData(userID, key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.states[userID]
	now := time.Now()
	if !exists || (m.ttl > 0 && now.After(item.expiresAt)) {
		item = stateItem{state: ""}
	}

	if item.data == nil {
		item.data = make(map[string]any)
	}
	item.data[key] = value
	if m.ttl > 0 {
		item.expiresAt = now.Add(m.ttl)
	}
	m.states[userID] = item
}

// GetData retrieves an arbitrary key-value pair for a specific user.
func (m *MemoryStorage) GetData(userID, key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.states[userID]
	if !exists || item.data == nil {
		return nil, false
	}

	if m.ttl > 0 && time.Now().After(item.expiresAt) {
		return nil, false
	}

	val, ok := item.data[key]
	return val, ok
}
