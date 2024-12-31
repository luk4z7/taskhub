package task

import (
	"context"

	"github.com/luk4z7/taskmanager/domain/user"
)

//go:generate mockgen -destination=mock/mock_task_repository.go -package=mock . TaskRepository
type TaskRepository interface {
	AddTask(ctx context.Context, payload Task) error
	List(ctx context.Context, role user.Role) ([]Task, error)
}
