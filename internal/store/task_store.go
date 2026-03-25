package store

import (
	"database/sql"
	"errors"
	"sync"
	"task-manager/internal/models"
	"time"

	"github.com/jmoiron/sqlx"
)

// sentinel error for not found
var ErrTaskNotFound = errors.New("task not found")

type MemoryTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]models.Task
}

type PgTaskStore struct {
	db *sqlx.DB
}

func NewPgTaskStore(db *sqlx.DB) *PgTaskStore {
	return &PgTaskStore{db: db}
}

func NewMemoryTaskStore() *MemoryTaskStore {
	return &MemoryTaskStore{
		tasks: make(map[string]models.Task),
	}
}

type TaskStore interface {
	Create(task *models.Task) error
	Get(id string) (*models.Task, error)
	GetAll() ([]models.Task, error)
	Update(id string, task *models.UpdateTaskInput) (*models.Task, error)
	Delete(id string) error
}

func (s *MemoryTaskStore) Create(task *models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.CreatedAt = time.Now()
	s.tasks[task.ID] = *task
	return nil
}

func (s *PgTaskStore) Create(task *models.Task) error {
	task.CreatedAt = time.Now()
	_, err := s.db.NamedExec("INSERT INTO tasks (id, title, status, created_at) VALUES (:id, :title, :status, :created_at)", &task)
	if err != nil {
		return err
	}
	return nil
}

func (s *MemoryTaskStore) Get(id string) (*models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	if !ok {
		return nil, ErrTaskNotFound
	}
	return &task, nil
}

func (s *PgTaskStore) Get(id string) (*models.Task, error) {
	var task models.Task
	err := s.db.Get(&task, "SELECT * FROM tasks WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *MemoryTaskStore) GetAll() ([]models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	taskList := make([]models.Task, 0, len(s.tasks))

	for _, task := range s.tasks {
		taskList = append(taskList, task)
	}

	return taskList, nil
}

func (s *PgTaskStore) GetAll() ([]models.Task, error) {
	var tasks []models.Task
	err := s.db.Select(&tasks, "SELECT * FROM tasks ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *MemoryTaskStore) Update(id string, task *models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.tasks[id]
	if !ok {
		return ErrTaskNotFound
	}
	now := time.Now()
	task.UpdatedAt = &now
	s.tasks[id] = *task

	return nil
}

func (s *PgTaskStore) Update(id string, taskToUpdate *models.UpdateTaskInput) (*models.Task, error) {
	now := time.Now()
	var task models.Task

	err := s.db.QueryRowx(`
		UPDATE tasks
		SET
			title = COALESCE($1, title),
			status = COALESCE($2, status),
			updated_at = $3
		WHERE id = $4
		RETURNING *
	`, taskToUpdate.Title, taskToUpdate.Status, now, id).StructScan(&task)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *MemoryTaskStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.tasks[id]
	if !ok {
		return ErrTaskNotFound
	}

	delete(s.tasks, id)
	return nil
}

func (s *PgTaskStore) Delete(id string) error {
	result, err := s.db.Exec("DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return err
	}

	numRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if numRows == 0 {
		return ErrTaskNotFound
	}
	return nil
}
