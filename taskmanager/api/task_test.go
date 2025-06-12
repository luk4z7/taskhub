package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"errors"
	"github.com/luk4z7/taskmanager/domain/task"
	mocktask "github.com/luk4z7/taskmanager/domain/task/mock"
	"github.com/luk4z7/taskmanager/domain/user"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"time"
)

type Server struct {
	recorder *httptest.ResponseRecorder
	req      *http.Request
	router   *echo.Echo
}

func setup(method, path string, body io.Reader, repo task.TaskRepository) Server {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")

	return Server{
		recorder: httptest.NewRecorder(),
		req:      req,
		router:   NewHttpRouter(nil, repo),
	}
}

const payload = `{ "summary": "hello world" }`

func TestCreateTask(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	taskMock := mocktask.NewMockTaskRepository(ctrl)
	taskMock.EXPECT().AddTask(context.Background(), gomock.Any()).Return(nil)

	server := setup(http.MethodPost, "/task", strings.NewReader(payload), taskMock)

	server.req.Header.Set("Authorization", "admin@domain.com")
	server.req.Header.Set("X-Role", "manager")
	ctx := server.router.NewContext(server.req, server.recorder)

	handler := Handler{
		task: task.New(taskMock),
	}

	err := handler.TaskSave(ctx)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, server.recorder.Code)

	uid, err := uuid.Parse(strings.Trim(server.recorder.Body.String(), "\n"))
	assert.Nil(t, err)
	t.Log(uid)
}

func TestTaskList(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskMock := mocktask.NewMockTaskRepository(ctrl)
	handler := Handler{
		task: task.New(taskMock),
	}

	t.Run("Success case - returns list of tasks", func(t *testing.T) {
		now := time.Now().UTC()
		user1 := user.User("user1@example.com")
		user2 := user.User("user2@example.com")

		expectedDomainTasks := []task.Task{
			task.MarshalTask("Task 1 summary", now, user1),
			task.MarshalTask("Task 2 summary", now.Add(-time.Hour), user2),
		}
		taskMock.EXPECT().List(gomock.Any(), gomock.Any()).Return(expectedDomainTasks, nil)

		server := setup(http.MethodGet, "/tasks", nil, taskMock)
		server.req.Header.Set("Authorization", "admin@domain.com")
		server.req.Header.Set("X-Role", "manager")
		ctx := server.router.NewContext(server.req, server.recorder)

		err := handler.TaskList(ctx)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, server.recorder.Code)

		// Construct expected JSON based on api.TaskList struct
		expectedJSON := `[
			{"summary":"Task 1 summary","created_at":"` + now.Format(time.RFC3339Nano) + `","created_by":"user1@example.com"},
			{"summary":"Task 2 summary","created_at":"` + now.Add(-time.Hour).Format(time.RFC3339Nano) + `","created_by":"user2@example.com"}
		]`
		assert.JSONEq(t, expectedJSON, server.recorder.Body.String())
	})

	t.Run("Error case - repository returns error", func(t *testing.T) {
		taskMock.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, errors.New("repository error"))

		server := setup(http.MethodGet, "/tasks", nil, taskMock)
		server.req.Header.Set("Authorization", "admin@domain.com")
		server.req.Header.Set("X-Role", "manager")
		ctx := server.router.NewContext(server.req, server.recorder)

		err := handler.TaskList(ctx)
		// The handler itself returns the error, Echo's default error handler will send the HTTP response.
		// We should assert that our handler did return an error.
		assert.Error(t, err)
		// Optionally, check the recorder if the error handler in Echo behaves predictably for this error.
		// For example, if it results in a 500:
		// assert.Equal(t, http.StatusInternalServerError, server.recorder.Code)
	})

	t.Run("Empty list case - repository returns empty list", func(t *testing.T) {
		expectedDomainTasks := []task.Task{}
		taskMock.EXPECT().List(gomock.Any(), gomock.Any()).Return(expectedDomainTasks, nil)

		server := setup(http.MethodGet, "/tasks", nil, taskMock)
		server.req.Header.Set("Authorization", "admin@domain.com")
		server.req.Header.Set("X-Role", "manager")
		ctx := server.router.NewContext(server.req, server.recorder)

		err := handler.TaskList(ctx)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, server.recorder.Code)
		assert.JSONEq(t, `[]`, server.recorder.Body.String()) // Empty array, no newline needed for JSONEq with empty array.
	})
}
