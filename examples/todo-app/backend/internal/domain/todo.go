package domain

import (
	"time"

	"github.com/google/uuid"
)

// Todo represents a todo item in the domain model
type Todo struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Priority    int       `json:"priority"`
	Tags        []string  `json:"tags"`
}

// CreateTodoRequest represents the request to create a new todo
type CreateTodoRequest struct {
	Title       string   `json:"title" validate:"required,min=1,max=200"`
	Description string   `json:"description" validate:"max=1000"`
	Priority    int      `json:"priority" validate:"gte=0,lte=5"`
	Tags        []string `json:"tags" validate:"max=10,dive,max=50"`
}

// UpdateTodoRequest represents the request to update an existing todo
type UpdateTodoRequest struct {
	Title       *string   `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string   `json:"description,omitempty" validate:"omitempty,max=1000"`
	Completed   *bool     `json:"completed,omitempty"`
	Priority    *int      `json:"priority,omitempty" validate:"omitempty,gte=0,lte=5"`
	Tags        *[]string `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`
}

// ListTodosRequest represents filtering and pagination for listing todos
type ListTodosRequest struct {
	Completed *bool  `json:"completed,omitempty"`
	Limit     int    `json:"limit" validate:"gte=1,lte=100"`
	Offset    int    `json:"offset" validate:"gte=0"`
	SortBy    string `json:"sort_by" validate:"omitempty,oneof=created_at updated_at priority title"`
	SortOrder string `json:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// TodoRepository defines the interface for todo storage operations
type TodoRepository interface {
	Create(todo *Todo) error
	GetByID(id uuid.UUID) (*Todo, error)
	List(req ListTodosRequest) ([]*Todo, int, error)
	Update(id uuid.UUID, req UpdateTodoRequest) (*Todo, error)
	Delete(id uuid.UUID) error
	HealthCheck() error
}

// TodoService provides business logic for todo operations
type TodoService struct {
	repo TodoRepository
}

// NewTodoService creates a new todo service
func NewTodoService(repo TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

// Create creates a new todo
func (s *TodoService) Create(req CreateTodoRequest) (*Todo, error) {
	now := time.Now()
	todo := &Todo{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
		Priority:    req.Priority,
		Tags:        req.Tags,
	}

	if err := s.repo.Create(todo); err != nil {
		return nil, err
	}

	return todo, nil
}

// GetByID retrieves a todo by ID
func (s *TodoService) GetByID(id uuid.UUID) (*Todo, error) {
	return s.repo.GetByID(id)
}

// List retrieves a list of todos with filtering and pagination
func (s *TodoService) List(req ListTodosRequest) ([]*Todo, int, error) {
	// Set defaults
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	return s.repo.List(req)
}

// Update updates an existing todo
func (s *TodoService) Update(id uuid.UUID, req UpdateTodoRequest) (*Todo, error) {
	return s.repo.Update(id, req)
}

// Delete deletes a todo
func (s *TodoService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

// HealthCheck verifies the service is healthy
func (s *TodoService) HealthCheck() error {
	return s.repo.HealthCheck()
}
