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
	Message string `json:"message"`
	ActorID string `json:"actor_id"`
}

type CacheRoomRequest struct {
	Actor       string `json:"actor"`
	ActorStatus bool   `json:"actorStatus"`
	RoomID      string `json:"roomId"`
}

type CacheRoomResponse struct {
	NewRoom      bool
	UserFullName string
	ToRef        string
	FromRef      string
	RoomIsLive   bool
}

type CacheChatRequest struct {
	ActorID string    `json:"actorId"`
	Read    bool      `json:"read"`
	ReadAt  time.Time `json:"readAt"`
	Message string    `json:"message"`
	RoomID  string    `json:"roomId"`
}

type GetCachedRequest struct {
	RoomId  string `bindQuery:"dataType=string" json:"roomId"`
	ActorId string
	Filter  *payload.GetListQueryNoPeriod `bindQuery:"dive=true"`
}
