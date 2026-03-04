//go:generate stringer -type=ChatEvent
package payload

import "time"

type ChatEvent int

const (
	NotifyJoin ChatEvent = iota
	NotifyLeave
	NotifyOnline
	NotifyOffline
	Message
	AlertError
)

type ChatMessage struct {
	Message string `json:"message"`
	FromID  string `json:"fromId"`
	ToID    string `json:"toId"`
	RoomID  string `json:"roomId"`
}

func (ctrl *ChatMessage) From(arg ...any) {

}

type AckPayload struct {
	Ack  bool      `json:"ack"`
	Time time.Time `json:"time"`
}
