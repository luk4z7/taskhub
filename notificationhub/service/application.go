package service

import (
	"context"
	"os"

	"github.com/luk4z7/notificationhub/app"
	"github.com/luk4z7/notificationhub/app/command"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func NewApplication(
	ctx context.Context,
	router *message.Router,
	logger watermill.LoggerAdapter,

) (app.Application, func()) {

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	ep, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return redisstream.NewSubscriber(redisstream.SubscriberConfig{
					Client:        rdb,
					ConsumerGroup: "issue-receipt",
				}, logger)
			},
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return "NotificationConfirmed", nil
			},
			Marshaler: cqrs.JSONMarshaler{},
			Logger:    logger,
		},
	)
	if err != nil {
		panic(err)
	}

	application := app.Application{
		Commands: app.Commands{
			Print: command.NewPrintHandler("PrintNotification"),
		},
	}

	if err := ep.AddHandlers(application.Commands.Print); err != nil {
		panic(err)
	}

	return application, func() {}
}
