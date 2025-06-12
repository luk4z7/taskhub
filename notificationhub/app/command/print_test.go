package command_test

import (
	// "bytes" // Removed unused import
	"context"
	"encoding/json"
	// "errors" // Removed unused import
	// "fmt" // Removed unused import
	"io"
	"log"
	"os"
	"strings"
	"testing"
	"time" // Added for PublishedAt

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/luk4z7/messages"
	"github.com/luk4z7/notificationhub/app/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPrintHandler(t *testing.T) {
	handlerName := "TestEmailSender"
	handler := command.NewPrintHandler(handlerName)

	require.NotNil(t, handler)
	// Internal field `name` is not exported, so we can't directly check it here
	// without also testing HandlerName(). We'll test HandlerName() separately.
	// We can assert that the type is correct.
	assert.IsType(t, &command.PrintHandler{}, handler)
}

func TestPrintHandler_HandlerName(t *testing.T) {
	expectedName := "MyTestHandler"
	handler := command.NewPrintHandler(expectedName)
	assert.Equal(t, expectedName, handler.HandlerName())
}

func TestPrintHandler_NewEvent(t *testing.T) {
	handler := command.NewPrintHandler("test")
	event := handler.NewEvent()
	assert.NotNil(t, event)
	assert.IsType(t, &message.Message{}, event)
}

// Helper function to capture stdout
func captureStdout(f func()) string {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old // restoring the real stdout
	return string(out)
}

func TestPrintHandler_Handle_Success(t *testing.T) {
	handler := command.NewPrintHandler("testHandler")
	ctx := context.Background()

	notificationPayload := messages.PrintNotification{
		Header:  messages.Header{ID: "msg-id-123", PublishedAt: time.Now().Format(time.RFC3339)}, // Corrected Header init
		Message: "Hello, this is a test notification!",
		Owner:   "test-owner",
	}
	payloadBytes, err := json.Marshal(notificationPayload)
	require.NoError(t, err)

	msg := message.NewMessage("test-uuid", payloadBytes)

	// var output string // Removed unused variable 'output'

	// Construct the expected string output from fmt.Println(notificationPayload) - This variable is also unused with current assertion strategy.
	// expectedOutput := fmt.Sprintf("%+v", notificationPayload)

	log.SetOutput(io.Discard) // Suppress log output from other parts if any, during capture
	defer log.SetOutput(os.Stderr) // Restore log output

	capturedOutput := captureStdout(func() {
		err = handler.Handle(ctx, msg)
	})
	assert.NoError(t, err)

	// Trim newline that fmt.Println adds, and any extra spaces from capture
	trimmedOutput := strings.TrimSpace(capturedOutput)
	assert.Contains(t, trimmedOutput, notificationPayload.Message, "Captured output should contain the message")
	assert.Contains(t, trimmedOutput, notificationPayload.Owner, "Captured output should contain the owner")
	assert.Contains(t, trimmedOutput, notificationPayload.Header.ID, "Captured output should contain the header ID")
	assert.Contains(t, trimmedOutput, notificationPayload.Header.PublishedAt, "Captured output should contain the PublishedAt timestamp")
}

func TestPrintHandler_Handle_NotAMessage(t *testing.T) {
	handler := command.NewPrintHandler("testHandler")
	ctx := context.Background()
	notAMessage := "this is a string, not a *message.Message"

	err := handler.Handle(ctx, notAMessage)
	require.Error(t, err)
	assert.EqualError(t, err, "this is not a *message.Message")
}

func TestPrintHandler_Handle_UnmarshalError(t *testing.T) {
	handler := command.NewPrintHandler("testHandler")
	ctx := context.Background()

	invalidPayload := []byte("this is not valid json")
	msg := message.NewMessage("test-uuid", invalidPayload)

	err := handler.Handle(ctx, msg)
	require.Error(t, err)
	// Check that the error is a json unmarshaling error
	// Example: "json: cannot unmarshal string into Go value of type messages.PrintNotification"
	// We can check for a substring.
	assert.Contains(t, err.Error(), "json: cannot unmarshal", "Error should be a JSON unmarshaling error")
}
