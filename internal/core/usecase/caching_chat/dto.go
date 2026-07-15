package caching_chat

import (
	"time"

	"github.com/rdhmuhammad/phisiobook/shared/payload"
)

type ChatRoomListResponse struct {
	RoomID         string `json:"roomId"`
	UserRefID      string `json:"userRefId"`
	LatestMessage  string `json:"latestMessage"`
	NotifCount     int    `json:"notifCount"`
	ProfilePicture string `json:"profilePicture"`
	Name           string `json:"name"`
}

type CachedChat struct {
	Message string    `json:"message"`
	ActorID string    `json:"actorId"`
	SendAt  time.Time `json:"sendAt"`
}

type CacheRoomRequest struct {
	Actor       string `json:"actor" example:"user_001"`
	ActorStatus bool   `json:"actorStatus" example:"true"`
	RoomID      string `json:"roomId" example:"room_abc123"`
}

type CacheRoomResponse struct {
	NewRoom      bool
	UserFullName string
	ToRef        string
	FromRef      string
	RoomIsLive   bool
}

type CacheChatRequest struct {
	ActorID string    `json:"actorId" example:"user_001"`
	Read    bool      `json:"read" example:"true"`
	ReadAt  time.Time `json:"readAt" example:"2024-01-15T10:00:00Z"`
	Message string    `json:"message" example:"Hello, how are you?"`
	RoomID  string    `json:"roomId" example:"room_abc123"`
}

type GetCachedRequest struct {
	RoomId  string `bindQuery:"dataType=string" json:"roomId" example:"room_abc123"`
	ActorId string
	Filter  *payload.GetListQueryNoPeriod `bindQuery:"dive=true"`
}
