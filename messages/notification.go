package messages

type PrintNotification struct {
	Header  Header `json:"header"`
	ID      string `json:"id"`
	Message any    `json:"message"`
	Owner   string `json:"owner"`
}
