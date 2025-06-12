package event

import (
	"context"
	"encoding/json"
	"errors"
	"log" // Standard Go log package
	"sync" // Added for sync.WaitGroup
	"time"

	"github.com/luk4z7/messages"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	// "github.com/sirupsen/logrus" // Removed
)

type (
	Worker struct {
		queue     chan messages.Message
		publisher *redisstream.Publisher
		// logger    *log.WatermillLogrusAdapter // Removed
		router   *message.Router
		eventBus *cqrs.EventBus
	}
)

// logging parameter removed
func NewWorker(publisher *redisstream.Publisher, router *message.Router) *Worker {
	eventBus, err := NewEventBus(publisher)
	if err != nil {
		panic(err) // This panic needs to be tested or handled more gracefully.
	}

	return &Worker{
		queue:     make(chan messages.Message, 10000),
		publisher: publisher,
		// logger:    logging, // Removed
		router:   router,
		eventBus: eventBus,
	}
}

func (w *Worker) Send(msg ...messages.Message) {
	for _, m := range msg {
		w.queue <- m
	}
}

func NewEventBus(pub message.Publisher) (*cqrs.EventBus, error) {
	// Use Watermill's standard logger which wraps Go's log package
	wmLogger := watermill.NewStdLogger(false, false) // debug=false, trace=false
	cqrsMarshaler := cqrs.JSONMarshaler{}

	if pub == nil { // Explicitly return an error to make NewWorker's panic testable
		return nil, errors.New("publisher cannot be nil for NewEventBus")
	}

	// cqrs.NewEventBusWithConfig will panic if pub is nil.
	// The panic message is "publisher is nil".
	return cqrs.NewEventBusWithConfig(pub, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			data, ok := params.Event.(*message.Message)
			if !ok {
				return "", errors.New("this is not a *message.Message")
			}
			// we can also use topic per event type
			topicName := data.Metadata.Get("type")
			if topicName == "" {
				// Adding an error for missing type, as this was a silent failure point before
				return "", errors.New("event message metadata 'type' is missing or empty")
			}
			return topicName, nil
		},
		OnPublish: func(params cqrs.OnEventSendParams) error {
			// Using standard log package
			// Removed params.Topic from log to avoid "undefined" error, which might be due to env/version issues.
			log.Printf("Publishing event: %s (topic determined by GeneratePublishTopic)", params.EventName)
			params.Message.Metadata.Set("published_at", time.Now().String())
			return nil
		},
		Marshaler: cqrsMarshaler,
		Logger:    wmLogger, // Use the Watermill standard logger
	})
}

func (w *Worker) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer log.Println("Worker internal queue processing goroutine truly finished.") // Renamed for clarity

		for {
			select {
			case msg, ok := <-w.queue:
				if !ok {
					log.Println("Worker queue channel closed, exiting processing goroutine.")
					return
				}
				log.Printf("Processing message from queue. TracingID: %s, Data Type: %T", msg.TracingID, msg.Data)

				switch data := msg.Data.(type) {
				case messages.PrintNotification:
					payload, err := json.Marshal(data)
					if err != nil {
						log.Printf("ERROR: JSON marshal failed for PrintNotification. TracingID: %s, Error: %v", msg.TracingID, err)
						continue
					}

					watermillMsg := message.NewMessage(watermill.NewUUID(), payload)
					watermillMsg.Metadata.Set("tracing_id", msg.TracingID)
					watermillMsg.Metadata.Set("type", "NotificationConfirmed")

					if err := w.eventBus.Publish(ctx, watermillMsg); err != nil {
						log.Printf("ERROR: Failed to publish NotificationConfirmed event. TracingID: %s, Error: %v. Re-queueing.", msg.TracingID, err)
						// Check context before re-queueing to prevent sending to a closed queue if context is also done.
						select {
						case <-ctx.Done():
							log.Println("Context cancelled while attempting to re-queue failed publish. Message dropped.")
						default:
							w.Send(msg) // Re-queue
						}
					} else {
						log.Printf("Successfully published NotificationConfirmed event. TracingID: %s", msg.TracingID)
					}

				default:
					log.Printf("ERROR: Unknown message type in queue. TracingID: %s, Data: %+v", msg.TracingID, msg.Data)
				}
			case <-ctx.Done():
				log.Println("Worker context cancelled, exiting processing goroutine.")
				return
			}
		}
	}()

	// Use Watermill's standard logger for retry middleware as well
	retryLogger := watermill.NewStdLogger(false, false)
	w.router.AddMiddleware(middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second,
		Multiplier:      2,
		Logger:          retryLogger, // Use the Watermill standard logger
	}.Middleware)

	log.Println("Worker router starting...")
	routerErr := w.router.Run(ctx) // Store router error

	if routerErr != nil {
		log.Printf("Worker router finished with error: %v", routerErr)
	} else {
		log.Println("Worker router finished gracefully.")
	}

	log.Println("Signaling queue processing goroutine to stop by closing queue...")
	close(w.queue) // Close w.queue to ensure the for/select loop in the goroutine terminates

	log.Println("Waiting for queue processing goroutine to finish...")
	wg.Wait() // Wait for the goroutine to actually finish
	log.Println("All worker components shut down.")

	return routerErr // Return the error from router.Run (if any)
}

func (w *Worker) Router() *message.Router {
	return w.router
}
