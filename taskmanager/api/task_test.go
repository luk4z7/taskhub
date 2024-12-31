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
	"github.com/luk4z7/taskmanager/domain/task"
	mocktask "github.com/luk4z7/taskmanager/domain/task/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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
