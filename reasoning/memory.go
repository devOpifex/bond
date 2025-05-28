package reasoning

import (
	"sync"
)

// Memory provides storage for state between reasoning steps
type Memory struct {
	mu    sync.RWMutex
	store map[string]interface{}
}

// NewMemory creates a new memory store
func NewMemory() *Memory {
	return &Memory{
		store: make(map[string]interface{}),
	}
}

// Get retrieves a value from memory
func (m *Memory) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, exists := m.store[key]
	return val, exists
}

// Set stores a value in memory
func (m *Memory) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[key] = value
}

// GetString retrieves a string value from memory
func (m *Memory) GetString(key string) (string, bool) {
	val, exists := m.Get(key)
	if !exists {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// GetAll returns all memory as a map
func (m *Memory) GetAll() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]interface{})
	for k, v := range m.store {
		result[k] = v
	}
	return result
}

// Clear removes all stored values
func (m *Memory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = make(map[string]interface{})
}