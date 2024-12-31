package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/luk4z7/taskmanager/domain/task"
	"github.com/luk4z7/taskmanager/domain/user"
)

const (
	dateFormatDefault = "2006-01-02 15:04:05"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}

func (t *TaskRepository) AddTask(ctx context.Context, payload task.Task) error {
	query := `INSERT INTO task (summary, created_at, created_by) VALUES (?, ?, ?)`

	stmt, err := t.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(
		payload.Summary,
		payload.CreatedAt(),
		payload.CreatedBy(),
	); err != nil {
		return err
	}

	return nil
}

func (t *TaskRepository) List(ctx context.Context, role user.Role) ([]task.Task, error) {
	var (
		tasks     []task.Task
		id        int64
		summary   string
		createdAt string
		createdBy string
	)

	sql := `SELECT task.id, summary, created_at, created_by FROM task`
	rows, err := t.db.Query(mountQuertByRole(sql, role))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&id,
			&summary,
			&createdAt,
			&createdBy,
		)
		if err != nil {
			return nil, err
		}

		t, err := time.Parse(dateFormatDefault, createdAt)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task.MarshalTask(
			summary,
			t,
			user.User(createdBy),
		))
	}

	return tasks, nil
}

func mountQuertByRole(sql string, role user.Role) string {
	if role != user.Manager {
		return fmt.Sprintf("%s %s", sql, ` INNER JOIN user ON (user.email = task.created_by) WHERE user.role = "`+role.String()+`"`)
	}

	return sql
}
