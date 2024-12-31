package messages

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	TracingID string
	Data      any
}

type Header struct {
	ID          string `json:"id"`
	PublishedAt string `json:"published_at"`
}

func NewHeader() Header {
	return Header{
		ID:          uuid.New().String(),
		PublishedAt: time.Now().Format(time.RFC3339),
	}
}

type EventHeader struct {
	ID          string    `json:"id"`
	PublishedAt time.Time `json:"published_at"`
}

func NewEventHeader() EventHeader {
	return EventHeader{
		ID:          uuid.NewString(),
		PublishedAt: time.Now().UTC(),
	}
}
