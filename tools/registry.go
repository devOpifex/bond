package tools

import (
	"fmt"
	"sync"

	"github.com/devOpifex/bond/models"
)

// Registry manages all available tools
type Registry struct {
	tools map[string]models.ToolExecutor
	mu    sync.RWMutex
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]models.ToolExecutor),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool models.ToolExecutor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := tool.GetName()
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool with name '%s' already registered", name)
	}
	
	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (models.ToolExecutor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	tool, exists := r.tools[name]
	return tool, exists
}

// GetAll returns all registered tools
func (r *Registry) GetAll() []models.ToolExecutor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	tools := make([]models.ToolExecutor, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Remove unregisters a tool
func (r *Registry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.tools, name)
}