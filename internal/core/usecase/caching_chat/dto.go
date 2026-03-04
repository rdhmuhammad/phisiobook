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
	ToID    string `json:"to_id"`
	Message string `json:"message"`
	FromID  string `json:"from_id"`
	RoomID  string `json:"room_id"`
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
	From    string    `json:"from"`
	Read    bool      `json:"read"`
	ReadAt  time.Time `json:"readAt"`
	To      string    `json:"to"`
	Message string    `json:"message"`
	RoomID  string    `json:"roomId"`
}

type GetCachedRequest struct {
	RoomId  string
	ActorId string
	Filter  *payload.GetListQueryNoPeriod
}
