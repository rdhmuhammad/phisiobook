//go:generate stringer -type=ChatEvent
package payload

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"encoding/json"

	"github.com/rdhmuhammad/phisiobook/pkg/logger"
)

type ChatEvent int

func (receiver ChatEvent) Topic() string {
	return strings.ToLower(receiver.String())
}

const (
	Notify_join ChatEvent = iota
	Notify_leave
	Notify_online
	Notify_offline
	Message
	Alert_error
)

type ChatMessage struct {
	ActorId string `json:"actorId"`
	Message string `json:"message"`
}

func (ctrl *ChatMessage) From(args ...any) {
	if ctrl == nil || len(args) == 0 {
		return
	}

	value := args[0]
	if nested, ok := value.([]any); ok && len(args) == 1 && len(nested) > 0 {
		value = nested[0]
	}

	switch msg := value.(type) {
	case ChatMessage:
		*ctrl = msg
		return
	case *ChatMessage:
		if msg != nil {
			*ctrl = *msg
		}
		return
	case string:
		st := cleanChatString(msg)
		if err := json.Unmarshal([]byte(st), ctrl); err != nil {
			ctrl.fromOrderedArgs(args)
		}
		return
	case []byte:
		st := cleanChatString(string(msg))
		if err := json.Unmarshal([]byte(st), ctrl); err != nil {
			ctrl.fromOrderedArgs(args)
		}
		return
	}

	if ctrl.fromMap(value) {
		return
	}

	raw, err := json.Marshal(value)
	if err == nil {
		_ = json.Unmarshal(raw, ctrl)
	}

	if len(args) > 1 {
		ctrl.fromOrderedArgs(args)
	}
}

func (ctrl *ChatMessage) fromMap(value any) bool {
	logger.Debugf("%v", value)
	val := reflect.ValueOf(value)
	if !val.IsValid() || val.Kind() != reflect.Map {
		return false
	}

	for _, key := range val.MapKeys() {
		if key.Kind() != reflect.String {
			continue
		}

		mapValue := val.MapIndex(key)
		if !mapValue.IsValid() || !mapValue.CanInterface() {
			continue
		}

		switch normalizeChatMessageKey(key.String()) {
		case "message":
			ctrl.Message = stringFromAny(mapValue.Interface())
		}
	}

	return true
}

func (ctrl *ChatMessage) fromOrderedArgs(args []any) {
	if len(args) > 0 && ctrl.Message == "" {
		ctrl.Message = stringFromAny(args[0])
	}
}

func cleanChatString(value string) string {
	value = strings.ReplaceAll(value, "\n", "")
	value = strings.ReplaceAll(value, "\r", "")
	return value
}

func normalizeChatMessageKey(key string) string {
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, "_", "")
	key = strings.ReplaceAll(key, "-", "")
	return key
}

func stringFromAny(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return cleanChatString(v)
	case []byte:
		return cleanChatString(string(v))
	default:
		return cleanChatString(fmt.Sprint(v))
	}
}

type AckPayload struct {
	Ack  bool      `json:"ack"`
	Time time.Time `json:"time"`
}
