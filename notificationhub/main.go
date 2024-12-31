package main

import (
	"context"

	"github.com/luk4z7/notificationhub/service"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()
	watermillLogger := log.NewWatermill(logrus.NewEntry(logrus.StandardLogger()))
	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(err)
	}

	_, cancelApp := service.NewApplication(ctx, router, watermillLogger)
	defer cancelApp()

	logrus.Info("router already and wating messages")
	router.Run(ctx)
}
