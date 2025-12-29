package caching_chat

import "base-be-golang/pkg/dto"

type CachedChat struct {
	ToID    string `json:"to_id"`
	Message string `json:"message"`
	RoomID  string `json:"room_id"`
}

type GetCachedRequest struct {
	RoomId  string
	ActorId string
	Filter  *dto.GetListQueryNoPeriod
}
