package api

import (
	"net/http"
	"time"

	"github.com/luk4z7/taskmanager/domain/task"
	"github.com/luk4z7/taskmanager/domain/user"

	"github.com/luk4z7/messages"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type (
	TaskRequest struct {
		Summary string `json:"summary"`
	}

	TaskList struct {
		Summary   string    `json:"summary"`
		CreatedAt time.Time `json:"created_at"`
		CreatedBy string    `json:"created_by"`
	}
)

func (h *Handler) TaskSave(e echo.Context) error {
	ctx := e.Request().Context()
	userRequest := e.Request().Header.Get("Authorization")
	role := e.Request().Header.Get("X-Role")

	req := TaskRequest{}
	if err := e.Bind(&req); err != nil {
		return err
	}

	if err := h.task.Save(ctx, task.Task{
		Summary: req.Summary,
	}, user.User(userRequest)); err != nil {
		return err
	}

	tracingID := uuid.New().String()

	if role == user.Technician.String() {
		h.w.Send(messages.Message{
			TracingID: tracingID,
			Data: messages.PrintNotification{
				Header:  messages.NewHeader(),
				Message: req.Summary,
				Owner:   userRequest,
			},
		})
	}

	return e.JSON(http.StatusCreated, tracingID)
}

func (h *Handler) TaskList(e echo.Context) error {
	ctx := e.Request().Context()

	taskList := []TaskList{}
	role := e.Request().Header.Get("X-Role")

	list, err := h.task.List(ctx, user.Role(role))
	if err != nil {
		return err
	}

	for _, v := range list {
		taskList = append(taskList, TaskList{
			Summary:   v.Summary,
			CreatedAt: v.CreatedAt(),
			CreatedBy: v.CreatedBy().String(),
		})
	}

	return e.JSON(http.StatusOK, taskList)
}
