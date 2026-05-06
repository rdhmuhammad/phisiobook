package socket

import (
	"context"

	"github.com/rdhmuhammad/phisiobook/internal/adapter/controller"
	"github.com/rdhmuhammad/phisiobook/internal/adapter/payload"
	"github.com/rdhmuhammad/phisiobook/internal/core/usecase/caching_chat"
	"github.com/rdhmuhammad/phisiobook/pkg/cio"
	"github.com/rdhmuhammad/phisiobook/pkg/logger"
	"github.com/rdhmuhammad/phisiobook/pkg/mongodb"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	"gorm.io/gorm"

	"time"

	"github.com/zishang520/socket.io/servers/socket/v3"
)

type ChatSocket struct {
	base.BaseSocket
	cachedUc controller.CachedChatUsecase
}

func NewChatSocket(dbConn *gorm.DB, mongoConn *mongodb.Conn, baseSocket base.BaseSocket, prt base.Port) ChatSocket {
	return ChatSocket{
		cachedUc:   caching_chat.NewUsecase(dbConn, mongoConn, prt),
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
		time.Now().Add(time.Second*time.Duration(ctrl.Env.GetInt("TIMEOUT_IN_SECOND", 5))),
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
		ctrl.Mapper.ErrorSocket(client, err)
		return
	}

	sockets := io.Space.To(targetRoom).FetchSockets()

	sockets(func(sockets []*socket.RemoteSocket, err error) {
		if err != nil {
			ctrl.ErrHandler.ErrorPrint(err)
			return
		}

		for _, sc := range sockets {
			if sc.Id() != client.Id() {
				if result.NewRoom && action == joining {
					err := sc.
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

				err = sc.
					Emit(actCacheEv[action].String(), payload.ChatMessage{
						Message: result.UserFullName + " Join the room",
						FromID:  result.FromRef,
						ToID:    result.ToRef,
					})
			}
		}
	})

}

func (ctrl ChatSocket) SendChat(io *cio.NS, client *socket.Socket, message cio.MessagePayload) {
	sentMsg := message.(*payload.ChatMessage)
	request := caching_chat.CacheChatRequest{
		From:    sentMsg.FromID,
		To:      sentMsg.ToID,
		Message: sentMsg.Message,
		RoomID:  sentMsg.RoomID,
	}
	tzName := client.Handshake().Query.Query().Get("tz")
	var tz = time.UTC
	var err error
	if tzName != "" {
		tz, err = time.LoadLocation(tzName)
		ctrl.ErrHandler.ErrorPrint(err)
	}

	remoteSocket := io.Space.To(socket.Room(sentMsg.RoomID)).FetchSockets()
	remoteSocket(func(sockets []*socket.RemoteSocket, err error) {
		for _, rs := range sockets {
			if rs.Id() != client.Id() {
				ack := io.Space.To(socket.Room(rs.Id())).
					Timeout(1*time.Second).
					EmitWithAck(payload.Message.String(), sentMsg)
				ack(func(response []any, err error) {
					if err != nil {
						request.Read = false
						return
					}
					request.Read = true
					request.ReadAt = time.Now().In(tz)
				})

				go ctrl.cachedUc.CacheChat(context.Background(), request)
			} else {
				err := io.Space.To(socket.Room(rs.Id())).
					Emit(payload.Message.String(), sentMsg)
				if err != nil {
					ctrl.Mapper.ErrorSocket(client, err)
					return
				}
			}
		}
	})

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
