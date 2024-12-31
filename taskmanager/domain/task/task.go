package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/luk4z7/taskmanager/domain/user"
)

const limitSummary = 2500

type Task struct {
	Summary   string
	createdAt time.Time
	createdBy user.User
}

func (t Task) CreatedAt() time.Time {
	return t.createdAt
}

func (t Task) CreatedBy() user.User {
	return t.createdBy
}

func MarshalTask(summary string, createdAt time.Time, user user.User) Task {
	return Task{
		Summary:   summary,
		createdAt: createdAt,
		createdBy: user,
	}
}

type TaskHandler struct {
	repo TaskRepository
}

func New(repo TaskRepository) *TaskHandler {
	return &TaskHandler{
		repo: repo,
	}
}

func (h *TaskHandler) Save(ctx context.Context, task Task, user user.User) error {
	if len(task.Summary) > limitSummary {
		return errors.New(fmt.Sprintf("summary exceded the limit of %d", limitSummary))
	}

	return h.repo.AddTask(ctx, Task{
		Summary:   task.Summary,
		createdAt: time.Now(),
		createdBy: user,
	})
}

func (h *TaskHandler) List(ctx context.Context, role user.Role) ([]Task, error) {
	return h.repo.List(ctx, role)
}
