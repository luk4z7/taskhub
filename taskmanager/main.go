package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/luk4z7/taskmanager/api"
	"github.com/luk4z7/taskmanager/db"

	"github.com/luk4z7/taskmanager/event"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func init() {
	log.Init(logrus.InfoLevel)
}

func publishers(rdb redis.UniversalClient, watermillLogger watermill.LoggerAdapter) *redisstream.Publisher {
	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	return publisher
}

func main() {
	watermillLogger := log.NewWatermill(logrus.NewEntry(logrus.StandardLogger()))

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	publisher := publishers(rdb, watermillLogger)

	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	w := event.NewWorker(watermillLogger, publisher, router)
	dbClient, err := db.MySqlHandler()
	if err != nil {
		panic(err)
	}

	if err := db.MigrationRun(ctx, dbClient); err != nil {
		panic(err)
	}

	e := api.NewHttpRouter(w, db.NewTaskRepository(dbClient))

	g.Go(func() error {
		return w.Run(ctx)
	})

	g.Go(func() error {
		logrus.Info("message router starting...")
		<-w.Router().Running()

		logrus.Info("Server starting...")

		err := e.Start(":8080")
		if err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	g.Go(func() error {
		// Shut down the HTTP server
		<-ctx.Done()

		logrus.Info("Server shutdown...")

		return e.Shutdown(ctx)
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-c
		e.Shutdown(ctx)
		os.Exit(2)
	}()

	// Will block until all goroutines finish
	if err := g.Wait(); err != nil {
		panic(err)
	}
}
