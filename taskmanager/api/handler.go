package api

import (
	"github.com/luk4z7/taskmanager/domain/task"
	"github.com/luk4z7/taskmanager/event"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type Handler struct {
	eventBus *cqrs.EventBus
	task     *task.TaskHandler
	w        *event.Worker
}
