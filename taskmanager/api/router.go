package api

import (
	"net/http"

	"github.com/luk4z7/taskmanager/domain/task"
	"github.com/luk4z7/taskmanager/event"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/common/http"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func NewHttpRouter(
	w *event.Worker,
	taskRepository task.TaskRepository,
) *echo.Echo {
	e := libHttp.NewEcho()

	e.Use(otelecho.Middleware("task"))

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	handler := Handler{
		w:    w,
		task: task.New(taskRepository),
	}

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.POST("/task", handler.TaskSave)
	e.GET("/task", handler.TaskList)

	return e
}
