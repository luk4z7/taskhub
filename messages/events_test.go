package messages_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/luk4z7/messages" // Import the package being tested
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHeader(t *testing.T) {
	header := messages.NewHeader()

	// Assert that Header.ID is a valid UUID string
	_, err := uuid.Parse(header.ID)
	assert.NoError(t, err, "ID should be a valid UUID")

	// Assert that Header.PublishedAt is a valid RFC3339 timestamp string
	parsedTime, err := time.Parse(time.RFC3339, header.PublishedAt)
	assert.NoError(t, err, "PublishedAt should be a valid RFC3339 timestamp string")

	// Check that the parsed time is recent (e.g., within a few seconds of time.Now())
	// Allow for a small delta, e.g., 5 seconds, to account for execution time.
	assert.WithinDuration(t, time.Now(), parsedTime, 5*time.Second, "PublishedAt should be a recent timestamp")
}

func TestNewEventHeader(t *testing.T) {
	eventHeader := messages.NewEventHeader()

	// Assert that EventHeader.ID is a valid UUID string
	_, err := uuid.Parse(eventHeader.ID)
	assert.NoError(t, err, "ID should be a valid UUID")

	// Assert that EventHeader.PublishedAt is a recent UTC timestamp
	require.NotNil(t, eventHeader.PublishedAt, "PublishedAt should not be nil")
	assert.Equal(t, time.UTC, eventHeader.PublishedAt.Location(), "PublishedAt should be in UTC")

	// Check that the timestamp is recent
	// Allow for a small delta, e.g., 5 seconds.
	assert.WithinDuration(t, time.Now().UTC(), eventHeader.PublishedAt, 5*time.Second, "PublishedAt should be a recent UTC timestamp")
}
