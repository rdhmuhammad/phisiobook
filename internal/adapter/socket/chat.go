package socket

import (
	"context"
	"strings"

	"github.com/rdhmuhammad/phisiobook/internal/adapter/controller"
	"github.com/rdhmuhammad/phisiobook/internal/adapter/payload"
	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/usecase/caching_chat"
	"github.com/rdhmuhammad/phisiobook/pkg/cio"
	"github.com/rdhmuhammad/phisiobook/pkg/localerror"
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
	joining: payload.Notify_join,
	leaving: payload.Notify_leave,
}

func (ctrl ChatSocket) cacheRoom(io *cio.NS, client *socket.Socket, action actionCache) {
	userRef := client.Handshake().Query.Query().Get("userRef")
	roomId := strings.TrimSpace(client.Handshake().Query.Query().Get("roomId"))
	targetRoom := socket.Room(roomId)
	client.Join(targetRoom)

	newContext, cancle := context.WithDeadline(
		context.Background(),
		time.Now().Add(time.Second*time.Duration(ctrl.Env.GetInt("TIMEOUT_IN_SECOND", 5))),
	)
	result, err := ctrl.cachedUc.CacheRoom(
		newContext,
		caching_chat.CacheRoomRequest{
			Actor:       userRef,
			ActorStatus: action == joining,
			RoomID:      roomId,
		},
		cancle,
	)
	if err != nil {
		logger.Error(err)
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
				ctrl.ErrHandler.DebugPrint("Client Socket ID => %s", client.Id())
				ctrl.ErrHandler.DebugPrint("Sc Socket ID => %s", sc.Id())
				if !result.NewRoom && action == joining {
					err := sc.
						Emit(payload.Notify_join.Topic(), payload.ChatMessage{
							Message: result.UserFullName + " Join the room",
						})
					if err != nil {
						logger.Error(err)
					}
					return
				}

				err = sc.
					Emit(actCacheEv[action].String(), payload.ChatMessage{
						Message: result.UserFullName + " Join the room",
					})
				if err != nil {
					logger.Error(err)
				}
				return
			}
		}
	})

}

func (ctrl ChatSocket) SendChat(io *cio.NS, client *socket.Socket, message cio.MessagePayload) {
	sentMsg := message.(*payload.ChatMessage)
	roomId := strings.TrimSpace(client.Handshake().Query.Query().Get("roomId"))
	if roomId == "" {
		ctrl.Mapper.ErrorSocket(client, localerror.InvalidData(constant.RoomNotValid.String()))
		return
	}

	userRef := strings.TrimSpace(client.Handshake().Query.Query().Get("userRef"))
	if userRef == "" {
		ctrl.Mapper.ErrorSocket(client, localerror.InvalidData(constant.AccessNotAllowed.String()))
		return
	}
	request := caching_chat.CacheChatRequest{
		ActorID: userRef,
		Message: sentMsg.Message,
		RoomID:  roomId,
	}
	tzName := client.Handshake().Query.Query().Get("tz")
	var tz = time.UTC
	if tzName != "" {
		if t, err := time.LoadLocation(tzName); err != nil {
			ctrl.ErrHandler.ErrorPrint(err)
		} else {
			tz = t
		}
	}

	remoteSocket := io.Space.To(socket.Room(roomId)).FetchSockets()
	remoteSocket(func(sockets []*socket.RemoteSocket, err error) {
		for _, rs := range sockets {
			if rs.Id() != client.Id() {
				ack := io.Space.To(socket.Room(rs.Id())).
					Timeout(1*time.Second).
					EmitWithAck(payload.Message.Topic(), sentMsg)
				ack(func(response []any, err error) {
					if err != nil {
						request.Read = false
						return
					}
					request.Read = true
					request.ReadAt = time.Now().In(tz)
				})

			} else {
				err := io.Space.To(socket.Room(rs.Id())).
					Emit(payload.Message.Topic(), sentMsg)
				if err != nil {
					ctrl.Mapper.ErrorSocket(client, err)
					return
				}
			}
		}
		go ctrl.cachedUc.CacheChat(context.Background(), request)
	})

}

func (ctrl ChatSocket) LeaveRoom(io *cio.NS, client *socket.Socket) {
	ctrl.cacheRoom(io, client, leaving)
}

func (ctrl ChatSocket) OnSpace(ns cio.NSInitiate) {
	ns("chat", nil).
		UserRoom().
		Auth(ctrl.Security.SocketValidate("token")).
		Connect(ctrl.JoinRoom).
		Event(payload.Message.Topic(), &payload.ChatMessage{}, ctrl.SendChat).
		Disconnect(ctrl.LeaveRoom).Build()
}
