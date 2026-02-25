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
	Update(id string, task *models.Task) error
	Delete(id string) error
}

func (s *MemoryTaskStore) Create(task *models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.CreatedAt = time.Now()
	s.tasks[task.ID] = *task
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
	if err == sql.ErrNoRows {
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

func (s *MemoryTaskStore) Update(id string, task *models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.tasks[id]
	if !ok {
		return ErrTaskNotFound
	}
	task.UpdatedAt = time.Now()
	s.tasks[id] = *task

	return nil
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
