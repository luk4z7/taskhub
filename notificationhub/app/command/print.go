package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/luk4z7/messages"

	"github.com/ThreeDotsLabs/watermill/message"
)

func NewPrintHandler(name string) *PrintHandler {
	return &PrintHandler{
		name: name,
	}
}

type PrintHandler struct {
	name string
}

func (p *PrintHandler) HandlerName() string {
	return p.name
}

func (p *PrintHandler) NewEvent() interface{} {
	return &message.Message{}
}

func (t *PrintHandler) Handle(ctx context.Context, event any) error {
	msg, ok := event.(*message.Message)
	if !ok {
		return errors.New("this is not a *message.Message")
	}

	var data messages.PrintNotification
	if err := json.Unmarshal(msg.Payload, &data); err != nil {
		return err
	}

	fmt.Println(data)

	return nil
}
