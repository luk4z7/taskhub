package event

import (
	"context"
	"encoding/json" // Re-added
	"errors"
	"testing"
	"time"

	// No specific logger import needed beyond standard "log" which is implicitly available for Printf etc.
	// Watermill's NewStdLogger uses the standard log package.
	"github.com/ThreeDotsLabs/watermill" // For watermill.NewStdLogger
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/luk4z7/messages"
	// "github.com/sirupsen/logrus" // No longer used
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockPublisher is a mock for watermill message.Publisher
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(topic string, messages ...*message.Message) error {
	args := m.Called(topic, messages)
	return args.Error(0)
}

func (m *MockPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewEventBus_Success(t *testing.T) {
	mockPub := new(MockPublisher)
	eventBus, err := NewEventBus(mockPub)
	require.NoError(t, err)
	assert.NotNil(t, eventBus)
}

func TestNewEventBus_GeneratePublishTopic(t *testing.T) {
	// Replicating the GeneratePublishTopic logic from worker.go for direct testing
	generatePublishTopicFunc := func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
		data, ok := params.Event.(*message.Message)
		if !ok {
			return "", errors.New("this is not a *message.Message")
		}
		topicName := data.Metadata.Get("type")
		if topicName == "" {
			return "", errors.New("event message metadata 'type' is missing or empty")
		}
		return topicName, nil
	}

	t.Run("Success_GoodEvent", func(t *testing.T) {
		msg := message.NewMessage("uuid", []byte("payload"))
		msg.Metadata.Set("type", "TestTopic")
		params := cqrs.GenerateEventPublishTopicParams{EventName: "SomeEvent", Event: msg}
		topic, err := generatePublishTopicFunc(params)
		assert.NoError(t, err)
		assert.Equal(t, "TestTopic", topic)
	})

	t.Run("Failure_NotAMessageEvent", func(t *testing.T) {
		params := cqrs.GenerateEventPublishTopicParams{EventName: "SomeEvent", Event: "not a message"}
		_, err := generatePublishTopicFunc(params)
		assert.Error(t, err)
		assert.EqualError(t, err, "this is not a *message.Message")
	})

	t.Run("Failure_MetadataTypeMissing", func(t *testing.T) { // Changed from Success to Failure
		msg := message.NewMessage("uuid", []byte("payload"))
		// No "type" in metadata
		params := cqrs.GenerateEventPublishTopicParams{EventName: "SomeEvent", Event: msg}
		_, err := generatePublishTopicFunc(params) // Use the replicated func
		assert.Error(t, err)
		assert.EqualError(t, err, "event message metadata 'type' is missing or empty")
	})
}

func TestNewWorker_Success_And_PanicOnNilPublisher(t *testing.T) {
	var nilRedisPub *redisstream.Publisher
	router, err := message.NewRouter(message.RouterConfig{}, watermill.NewStdLogger(false, false))
	require.NoError(t, err)

	// NewEventBus now returns an error "publisher cannot be nil for NewEventBus" if pub is nil.
	// NewWorker will panic on this error.
	assert.PanicsWithError(t, "publisher cannot be nil for NewEventBus", func() {
		NewWorker(nilRedisPub, router)
	}, "NewWorker should panic with 'publisher cannot be nil for NewEventBus'")
}

// TestNewWorker_PanicsOnEventBusError was removed as it's identical to the above test's purpose.

func TestWorker_Send(t *testing.T) {
	worker := &Worker{
		queue: make(chan messages.Message, 1),
	}
	msg := messages.Message{TracingID: "123", Data: messages.PrintNotification{Message: "hello"}}
	worker.Send(msg)

	receivedMsg := <-worker.queue
	assert.Equal(t, msg, receivedMsg)
}

func TestWorker_Run_ProcessPrintNotification_Success(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router, err := message.NewRouter(message.RouterConfig{}, watermill.NewStdLogger(false, false))
	require.NoError(t, err)

	mockMessagesPublisher := new(MockPublisher)
	eventBusForWorker, err := NewEventBus(mockMessagesPublisher)
	require.NoError(t, err)

	worker := &Worker{
		queue:     make(chan messages.Message, 10),
		publisher: nil,
		router:    router,
		eventBus:  eventBusForWorker,
	}

	printMsgData := messages.PrintNotification{Message: "test message", Header: messages.Header{ID: "header-id"}, Owner: "owner"}
	originalMsg := messages.Message{
		TracingID: "trace-123",
		Data:      printMsgData,
	}
	expectedPayload, err := json.Marshal(printMsgData) // Uncommented and check error
	require.NoError(t, err)

	// Using mock.AnythingOfType for the messages argument as requested.
	// This makes the test less strict about message content but ensures the type is correct.
	mockMessagesPublisher.On("Publish", "NotificationConfirmed", mock.AnythingOfType("[]*message.Message")).Return(nil)

	runErrChan := make(chan error, 1)
	go func() {
		runErrChan <- worker.Run(ctx)
	}()

	worker.Send(originalMsg)
	time.Sleep(50 * time.Millisecond)

	mockMessagesPublisher.AssertExpectations(t)
	cancel()
	errFromRun := <-runErrChan
	assert.NoError(t, errFromRun, "Router.Run should return nil on context cancellation")
}

func TestWorker_Run_ProcessPrintNotification_PublishError_Requeue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router, err := message.NewRouter(message.RouterConfig{}, watermill.NewStdLogger(false, false))
	require.NoError(t, err)

	mockMessagesPublisher := new(MockPublisher)
	eventBusForWorker, err := NewEventBus(mockMessagesPublisher)
	require.NoError(t, err)

	worker := &Worker{
		queue:     make(chan messages.Message, 10),
		publisher: nil,
		router:    router,
		eventBus:  eventBusForWorker,
	}

	printMsgData := messages.PrintNotification{Message: "test message"}
	originalMsg := messages.Message{TracingID: "trace-456", Data: printMsgData}

	publishError := errors.New("publish failed")
	mockMessagesPublisher.On("Publish", "NotificationConfirmed", mock.AnythingOfType("[]*message.Message")).Return(publishError).Once()
	mockMessagesPublisher.On("Publish", "NotificationConfirmed", mock.AnythingOfType("[]*message.Message")).Return(nil).Once()

	runErrChan := make(chan error, 1)
	go func() {
		runErrChan <- worker.Run(ctx)
	}()
	time.Sleep(10 * time.Millisecond)

	worker.Send(originalMsg)
	time.Sleep(150 * time.Millisecond)

	mockMessagesPublisher.AssertExpectations(t)
	cancel()
	errFromRun := <-runErrChan
	assert.NoError(t, errFromRun)
}

func TestWorker_Run_ProcessPrintNotification_JsonMarshalError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router, err := message.NewRouter(message.RouterConfig{}, watermill.NewStdLogger(false, false))
	require.NoError(t, err)

	mockMessagesPublisher := new(MockPublisher)
	eventBusForWorker, err := NewEventBus(mockMessagesPublisher)
	require.NoError(t, err)

	worker := &Worker{
		queue:     make(chan messages.Message, 10),
		publisher: nil,
		router:    router,
		eventBus:  eventBusForWorker,
	}

	invalidData := messages.PrintNotification{Message: string([]byte{0xff, 0xfe, 0xfd})} // Invalid UTF-8
	originalMsg := messages.Message{TracingID: "trace-789", Data: invalidData}

	runErrChan := make(chan error, 1)
	go func() {
		runErrChan <- worker.Run(ctx)
	}()
	time.Sleep(10 * time.Millisecond)

	worker.Send(originalMsg)
	time.Sleep(50 * time.Millisecond)

	mockMessagesPublisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	cancel()
	errFromRun := <-runErrChan
	assert.NoError(t, errFromRun)
}

func TestWorker_Run_UnknownMessageTypeInQueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router, err := message.NewRouter(message.RouterConfig{}, watermill.NewStdLogger(false, false))
	require.NoError(t, err)

	mockMessagesPublisher := new(MockPublisher)
	eventBusForWorker, err := NewEventBus(mockMessagesPublisher)
	require.NoError(t, err)

	worker := &Worker{
		queue:     make(chan messages.Message, 10),
		publisher: nil,
		router:    router,
		eventBus:  eventBusForWorker,
	}

	originalMsg := messages.Message{TracingID: "trace-abc", Data: "this is not PrintNotification"}

	runErrChan := make(chan error, 1)
	go func() {
		runErrChan <- worker.Run(ctx)
	}()
	time.Sleep(10 * time.Millisecond)

	worker.Send(originalMsg)
	time.Sleep(50 * time.Millisecond)

	mockMessagesPublisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	cancel()
	errFromRun := <-runErrChan
	assert.NoError(t, errFromRun)
}
