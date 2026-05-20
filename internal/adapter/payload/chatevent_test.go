package payload

import "testing"

func TestChatMessageFromMapPayload(t *testing.T) {
	msg := ChatMessage{}
	msg.From(map[string]any{
		"message": "hello",
		"fromId":  "user-1",
		"toId":    "user-2",
		"roomId":  "room-1",
	})

	assertChatMessage(t, msg)
}

func TestChatMessageFromNestedSocketArgs(t *testing.T) {
	msg := ChatMessage{}
	msg.From([]any{
		map[string]any{
			"message": "hello",
			"fromId":  "user-1",
			"toId":    "user-2",
			"roomId":  "room-1",
		},
	})

	assertChatMessage(t, msg)
}

func TestChatMessageFromJSONPayload(t *testing.T) {
	msg := ChatMessage{}
	msg.From(`{"message":"hello","fromId":"user-1","toId":"user-2","roomId":"room-1"}`)

	assertChatMessage(t, msg)
}

func TestChatMessageFromJSONPayloadWithNewlines(t *testing.T) {
	msg := ChatMessage{}
	msg.From("{\n\"message\":\"hello\",\r\n\"fromId\":\"user-1\",\"toId\":\"user-2\",\"roomId\":\"room-1\"\n}")

	assertChatMessage(t, msg)
}

func TestChatMessageFromOrderedArgs(t *testing.T) {
	msg := ChatMessage{}
	msg.From("hello", "user-1", "user-2", "room-1")

	assertChatMessage(t, msg)
}

func assertChatMessage(t *testing.T, msg ChatMessage) {
	t.Helper()

	if msg.Message != "hello" {
		t.Fatalf("Message = %q", msg.Message)
	}
	if msg.FromID != "user-1" {
		t.Fatalf("FromID = %q", msg.FromID)
	}
	if msg.ToID != "user-2" {
		t.Fatalf("ToID = %q", msg.ToID)
	}
	if msg.RoomID != "room-1" {
		t.Fatalf("RoomID = %q", msg.RoomID)
	}
}
