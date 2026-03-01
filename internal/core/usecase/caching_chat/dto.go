package caching_chat

import "base-be-golang/pkg/dto"

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
	From    string `json:"from"`
	To      string `json:"to"`
	Message string `json:"message"`
	RoomID  string `json:"roomId"`
}

type GetCachedRequest struct {
	RoomId  string
	ActorId string
	Filter  *dto.GetListQueryNoPeriod
}
