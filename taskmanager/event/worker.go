package event

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/luk4z7/messages"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/sirupsen/logrus"
)

type (
	Worker struct {
		queue     chan messages.Message
		publisher *redisstream.Publisher
		logger    *log.WatermillLogrusAdapter
		router    *message.Router
		eventBus  *cqrs.EventBus
	}
)

func NewWorker(logging *log.WatermillLogrusAdapter, publisher *redisstream.Publisher, router *message.Router) *Worker {
	eventBus, err := NewEventBus(publisher)
	if err != nil {
		panic(err)
	}

	return &Worker{
		queue:     make(chan messages.Message, 10000),
		publisher: publisher,
		logger:    logging,
		router:    router,
		eventBus:  eventBus,
	}
}

func (w *Worker) Send(msg ...messages.Message) {
	for _, m := range msg {
		w.queue <- m
	}
}

func NewEventBus(pub message.Publisher) (*cqrs.EventBus, error) {
	logger := watermill.NewStdLogger(false, false)
	cqrsMarshaler := cqrs.JSONMarshaler{}

	return cqrs.NewEventBusWithConfig(pub, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			data, ok := params.Event.(*message.Message)
			if !ok {
				return "", errors.New("this is not a *message.Message")
			}

			// we can also use topic per event type
			return data.Metadata["type"], nil
		},
		OnPublish: func(params cqrs.OnEventSendParams) error {
			logger.Info("Publishing event", watermill.LogFields{
				"event_name": params.EventName,
			})

			params.Message.Metadata.Set("published_at", time.Now().String())

			return nil
		},

		Marshaler: cqrsMarshaler,
		Logger:    logger,
	})
}

func (w *Worker) Run(ctx context.Context) error {
	go func() {
		for msg := range w.queue {
			logrus.Info("queue ", "data ", msg.Data)

			switch msg.Data.(type) {
			case messages.PrintNotification:

				payload, err := json.Marshal(msg.Data)
				if err != nil {
					logrus.Error(" wrong payload ", "data ", payload)
				}

				mesg := message.NewMessage(watermill.NewUUID(), payload)
				mesg.Metadata.Set("tracing_id", msg.TracingID)
				mesg.Metadata.Set("type", "NotificationConfirmed")

				if err := w.eventBus.Publish(ctx, mesg); err != nil {
					logrus.Error("error on publish confirmed ", "err ", err)
					w.Send(msg)
				}

			default:
				logrus.Error(" wrong queue ", "data ", msg.Data)
			}
		}
	}()

	w.router.AddMiddleware(middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second,
		Multiplier:      2,
		Logger:          w.logger,
	}.Middleware)

	return w.router.Run(ctx)
}

func (w *Worker) Router() *message.Router {
	return w.router
}
