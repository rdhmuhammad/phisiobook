package socket

import (
	"base-be-golang/internal/adapter/controller"
	"base-be-golang/internal/adapter/payload"
	"base-be-golang/internal/core/port"
	"base-be-golang/internal/core/usecase/caching_chat"
	"base-be-golang/pkg/cio"
	"base-be-golang/pkg/logger"
	"base-be-golang/pkg/mongodb"
	"context"
	"time"

	"github.com/zishang520/socket.io/servers/socket/v3"
)

type ChatSocket struct {
	BaseSocket
	cachedUc controller.CachedChatUsecase
}

func NewChatSocket(mongoConn *mongodb.Conn, baseSocket BaseSocket, prt port.Port) ChatSocket {
	return ChatSocket{
		cachedUc:   caching_chat.NewUsecase(mongoConn, prt),
		BaseSocket: baseSocket,
	}
}

func (ctrl ChatSocket) JoinRoom(io *cio.NS, client *socket.Socket) {
	ctrl.cacheRoom(io, client, joining)
}

type actionCache int

const (
	leaving actionCache = iota
	joining
)

var actCacheEv = map[actionCache]payload.ChatEvent{
	joining: payload.NotifyOnline,
	leaving: payload.NotifyOffline,
}

func (ctrl ChatSocket) cacheRoom(io *cio.NS, client *socket.Socket, action actionCache) {
	userRef := client.Handshake().Query.Query().Get("userRef")
	roomId := client.Handshake().Query.Query().Get("roomId")
	keys := client.Rooms().Keys()
	var targetRoom socket.Room
	for _, key := range keys {
		if roomId == string(key) {
			targetRoom = key
		}
	}
	newContext, cancle := context.WithDeadline(
		context.Background(),
		time.Now().Add(time.Second*time.Duration(ctrl.env.GetInt("TIMEOUT_IN_SECOND", 5))),
	)
	result, err := ctrl.cachedUc.CacheRoom(
		newContext,
		caching_chat.CacheRoomRequest{
			Actor:       userRef,
			ActorStatus: true,
			RoomID:      roomId,
		},
		cancle,
	)
	if err != nil {
		ctrl.mapper.ErrorSocket(client, err)
		return
	}

	if result.NewRoom && action == joining {
		err := io.Space.To(targetRoom).
			Emit(payload.NotifyJoin.String(), payload.ChatMessage{
				Message: result.UserFullName + " Join the room",
				FromID:  result.FromRef,
				ToID:    result.ToRef,
			})
		if err != nil {
			logger.Error(err)
		}
		return
	}

	err = io.Space.To(targetRoom).
		Emit(actCacheEv[action].String(), payload.ChatMessage{
			Message: result.UserFullName + " Join the room",
			FromID:  result.FromRef,
			ToID:    result.ToRef,
		})
}

func (ctrl ChatSocket) SendChat(io *cio.NS, client *socket.Socket, message cio.MessagePayload) {
	sentMsg := message.(*payload.ChatMessage)
	go ctrl.cachedUc.CacheChat(context.Background(), caching_chat.CacheChatRequest{
		From:    sentMsg.FromID,
		To:      sentMsg.ToID,
		Message: sentMsg.Message,
		RoomID:  sentMsg.RoomID,
	})

	err := io.Space.To(socket.Room(sentMsg.RoomID)).Emit(payload.Message.String(), sentMsg)
	if err != nil {
		ctrl.mapper.ErrorSocket(client, err)
		return
	}

}

func (ctrl ChatSocket) LeaveRoom(io *cio.NS, client *socket.Socket) {
	ctrl.cacheRoom(io, client, leaving)
}

func (ctrl ChatSocket) OnSpace(ns cio.NSInitiate) {
	ns("chat", nil).
		UserRoom().
		Connect(ctrl.JoinRoom).
		Event(payload.Message.String(), &payload.ChatMessage{}, ctrl.SendChat).
		Disconnect(ctrl.LeaveRoom).Build()
}
