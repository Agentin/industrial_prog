package repository

import (
	"context"

	"github.com/student/tech-ip-sem2/services/tasks/internal/service"
)

type TaskRepository interface {
	Create(ctx context.Context, task service.Task) error
	GetAll(ctx context.Context) ([]service.Task, error)
	GetByID(ctx context.Context, id string) (service.Task, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	Delete(ctx context.Context, id string) error
	SearchByTitle(ctx context.Context, title string) ([]service.Task, error) // для демонстрации SQLi
}
