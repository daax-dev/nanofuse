package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/marcboeker/go-duckdb"

	"github.com/daax-dev/nanofuse/examples/todo-app/backend/internal/domain"
)

var (
	ErrNotFound      = errors.New("todo not found")
	ErrAlreadyExists = errors.New("todo already exists")
)

// DuckDBRepository implements TodoRepository using DuckDB
type DuckDBRepository struct {
	db *sql.DB
}

// NewDuckDBRepository creates a new DuckDB repository
func NewDuckDBRepository(dbPath string) (*DuckDBRepository, error) {
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	repo := &DuckDBRepository{db: db}

	if err := repo.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

// initSchema creates the database schema if it doesn't exist
func (r *DuckDBRepository) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS todos (
		id VARCHAR PRIMARY KEY,
		title VARCHAR NOT NULL,
		description TEXT,
		completed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		priority INTEGER DEFAULT 0,
		tags VARCHAR[]
	);

	CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos(completed);
	CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos(created_at);
	CREATE INDEX IF NOT EXISTS idx_todos_priority ON todos(priority);
	`

	_, err := r.db.Exec(schema)
	return err
}

// Create creates a new todo
func (r *DuckDBRepository) Create(todo *domain.Todo) error {
	query := `
		INSERT INTO todos (id, title, description, completed, created_at, updated_at, priority, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Convert tags slice to DuckDB list format
	tagsStr := r.serializeTags(todo.Tags)

	_, err := r.db.Exec(query,
		todo.ID.String(),
		todo.Title,
		todo.Description,
		todo.Completed,
		todo.CreatedAt,
		todo.UpdatedAt,
		todo.Priority,
		tagsStr,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return ErrAlreadyExists
		}
		return fmt.Errorf("failed to create todo: %w", err)
	}

	return nil
}

// serializeTags converts a slice of tags to DuckDB list format
func (r *DuckDBRepository) serializeTags(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	// Create a DuckDB list literal: ['tag1', 'tag2']
	quoted := make([]string, len(tags))
	for i, tag := range tags {
		// Escape single quotes in tags
		escaped := strings.ReplaceAll(tag, "'", "''")
		quoted[i] = fmt.Sprintf("'%s'", escaped)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// deserializeTags converts DuckDB list to string slice
func (r *DuckDBRepository) deserializeTags(tagsStr string) []string {
	// Remove brackets and split
	tagsStr = strings.Trim(tagsStr, "[]")
	if tagsStr == "" {
		return []string{}
	}

	// Simple split by comma (assumes no commas in tags)
	parts := strings.Split(tagsStr, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, "'\"")
		if part != "" {
			tags = append(tags, part)
		}
	}
	return tags
}

// GetByID retrieves a todo by ID
func (r *DuckDBRepository) GetByID(id uuid.UUID) (*domain.Todo, error) {
	query := `
		SELECT id, title, description, completed, created_at, updated_at, priority, tags::VARCHAR
		FROM todos
		WHERE id = ?
	`

	var todo domain.Todo
	var idStr string
	var tagsStr string

	err := r.db.QueryRow(query, id.String()).Scan(
		&idStr,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.Priority,
		&tagsStr,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	todo.ID = id
	todo.Tags = r.deserializeTags(tagsStr)

	return &todo, nil
}

// List retrieves a list of todos with filtering and pagination
func (r *DuckDBRepository) List(req domain.ListTodosRequest) ([]*domain.Todo, int, error) {
	// Build query
	queryParts := []string{"SELECT id, title, description, completed, created_at, updated_at, priority, tags FROM todos"}
	countParts := []string{"SELECT COUNT(*) FROM todos"}
	args := []interface{}{}

	// Add WHERE clause if filtering by completion status
	if req.Completed != nil {
		queryParts = append(queryParts, "WHERE completed = ?")
		countParts = append(countParts, "WHERE completed = ?")
		args = append(args, *req.Completed)
	}

	// Add ORDER BY
	orderBy := fmt.Sprintf("ORDER BY %s %s", req.SortBy, strings.ToUpper(req.SortOrder))
	queryParts = append(queryParts, orderBy)

	// Add LIMIT and OFFSET
	queryParts = append(queryParts, "LIMIT ? OFFSET ?")
	query := strings.Join(queryParts, " ")
	queryArgs := append(args, req.Limit, req.Offset)

	// Get total count
	countQuery := strings.Join(countParts, " ")
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count todos: %w", err)
	}

	// Get todos - cast tags to VARCHAR for retrieval
	selectQuery := strings.Replace(query, "SELECT id, title, description, completed, created_at, updated_at, priority, tags FROM todos",
		"SELECT id, title, description, completed, created_at, updated_at, priority, tags::VARCHAR FROM todos", 1)
	rows, err := r.db.Query(selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list todos: %w", err)
	}
	defer rows.Close()

	todos := make([]*domain.Todo, 0)
	for rows.Next() {
		var todo domain.Todo
		var idStr string
		var tagsStr string

		err := rows.Scan(
			&idStr,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
			&todo.Priority,
			&tagsStr,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan todo: %w", err)
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse todo ID: %w", err)
		}

		todo.ID = id
		todo.Tags = r.deserializeTags(tagsStr)

		todos = append(todos, &todo)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating todos: %w", err)
	}

	return todos, total, nil
}

// Update updates an existing todo
func (r *DuckDBRepository) Update(id uuid.UUID, req domain.UpdateTodoRequest) (*domain.Todo, error) {
	// Build dynamic update query
	setParts := []string{"updated_at = ?"}
	args := []interface{}{time.Now()}

	if req.Title != nil {
		setParts = append(setParts, "title = ?")
		args = append(args, *req.Title)
	}
	if req.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *req.Description)
	}
	if req.Completed != nil {
		setParts = append(setParts, "completed = ?")
		args = append(args, *req.Completed)
	}
	if req.Priority != nil {
		setParts = append(setParts, "priority = ?")
		args = append(args, *req.Priority)
	}
	if req.Tags != nil {
		setParts = append(setParts, "tags = ?")
		args = append(args, r.serializeTags(*req.Tags))
	}

	args = append(args, id.String())

	query := fmt.Sprintf("UPDATE todos SET %s WHERE id = ?", strings.Join(setParts, ", "))

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, ErrNotFound
	}

	// Fetch and return the updated todo
	return r.GetByID(id)
}

// Delete deletes a todo
func (r *DuckDBRepository) Delete(id uuid.UUID) error {
	query := "DELETE FROM todos WHERE id = ?"

	result, err := r.db.Exec(query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// HealthCheck verifies the database connection is healthy
func (r *DuckDBRepository) HealthCheck() error {
	return r.db.Ping()
}

// Close closes the database connection
func (r *DuckDBRepository) Close() error {
	return r.db.Close()
}
