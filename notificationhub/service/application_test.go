package service_test

import (
	"context"
	"os"
	"testing"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/luk4z7/notificationhub/app"
	"github.com/luk4z7/notificationhub/service" // Import the package being tested
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApplication_Success(t *testing.T) {
	// Set required environment variable
	originalRedisAddr, redisAddrSet := os.LookupEnv("REDIS_ADDR")
	os.Setenv("REDIS_ADDR", "localhost:6379") // Actual connection not made by NewApplication directly for this part
	defer func() {
		if redisAddrSet {
			os.Setenv("REDIS_ADDR", originalRedisAddr)
		} else {
			os.Unsetenv("REDIS_ADDR")
		}
	}()

	logger := watermill.NewStdLogger(false, false) // debug=false, trace=false
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	require.NoError(t, err)

	var application app.Application
	var cleanupFunc func()

	assert.NotPanics(t, func() {
		application, cleanupFunc = service.NewApplication(context.Background(), router, logger)
	}, "NewApplication should not panic in the happy path")

	assert.NotNil(t, application, "Application should be initialized")
	assert.NotNil(t, application.Commands.Print, "Print command handler should be initialized")
	assert.Equal(t, "PrintNotification", application.Commands.Print.HandlerName(), "PrintHandler should have the correct name")
	assert.NotNil(t, cleanupFunc, "Cleanup function should be returned")

	// Execute cleanup func to ensure it doesn't panic (it's empty, but good practice)
	assert.NotPanics(t, cleanupFunc)
}

func TestNewApplication_PanicOnMissingRedisAddr(t *testing.T) {
	originalRedisAddr, redisAddrSet := os.LookupEnv("REDIS_ADDR")
	os.Unsetenv("REDIS_ADDR") // Ensure REDIS_ADDR is not set
	defer func() {
		if redisAddrSet {
			os.Setenv("REDIS_ADDR", originalRedisAddr)
		} else {
			// If it was originally unset, leave it unset.
			// However, to ensure safety for other tests if it was set by another parallel test,
			// explicitly unsetting it if it was not set before this test is fine.
			os.Unsetenv("REDIS_ADDR")
		}
	}()

	logger := watermill.NewStdLogger(false, false)
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	require.NoError(t, err)

	// NewApplication is expected to panic because redisstream.NewSubscriber
	// (called within SubscriberConstructor) will fail if REDIS_ADDR is empty.
	// NewApplication now has an explicit panic for this.
	assert.PanicsWithValue(t, "REDIS_ADDR environment variable is not set or is empty", func() {
		service.NewApplication(context.Background(), router, logger)
	}, "NewApplication should panic with specific message if REDIS_ADDR is not set")
}

func TestNewApplication_PanicOnRouterNil(t *testing.T) {
	// This test checks if NewEventProcessorWithConfig panics when router is nil
	originalRedisAddr, redisAddrSet := os.LookupEnv("REDIS_ADDR")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	defer func() {
		if redisAddrSet {
			os.Setenv("REDIS_ADDR", originalRedisAddr)
		} else {
			os.Unsetenv("REDIS_ADDR")
		}
	}()
	logger := watermill.NewStdLogger(false, false)

	// cqrs.NewEventProcessorWithConfig expects a non-nil router.
	// If router is nil, it should panic directly.
	// The test output indicates the panic value is indeed "missing router".
	// Let's use a more robust way to check this if PanicsWithValue is tricky.
	var recoveredValue interface{}
	func() {
		defer func() {
			recoveredValue = recover()
		}()
		service.NewApplication(context.Background(), nil, logger)
	}()
	require.NotNil(t, recoveredValue, "Expected a panic when router is nil")
	// Check that the panic value is an error and its message is "missing router"
	err, ok := recoveredValue.(error)
	require.True(t, ok, "Panic value should be an error type")
	assert.EqualError(t, err, "missing router", "Panic error message should be 'missing router'")
}

// Testing failure of ep.AddHandlers is harder without deeper mocking
// because command.NewPrintHandler("PrintNotification") always creates a valid handler.
// To make AddHandlers fail, we'd need to inject a faulty handler or mock the EventProcessor.
// For now, this path is considered less critical to unit test than the configuration/setup panics.
