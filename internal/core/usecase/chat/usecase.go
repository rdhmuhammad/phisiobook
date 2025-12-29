package chat

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/domain"
	"base-be-golang/internal/core/port"
	"base-be-golang/pkg/chat_io"
	"base-be-golang/pkg/localerror"
	"base-be-golang/pkg/mongodb"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
	"time"
)

type Usecase struct {
	port.Port
	roomRepo *mongodb.BaseRepo[domain.RoomSession]
	hub      *chat_io.Hub
}

func NewUsecase(mongoConn mongodb.Conn) Usecase {
	return Usecase{}
}

func (u Usecase) EnterRoom(c *gin.Context, roleName string, roomCode string, userCode string) error {
	ctx := c.Request.Context()
	roomSession, err := u.roomRepo.FindOne(ctx, "socket_room_id", roomCode)
	if err != nil {
		return err
	}

	if roomSession.ParticipantSocketIDEqual(roleName, userCode) {
		return localerror.AccessControlError{Msg: "Room session does not match"}
	}

	newContext, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*10))
	go func() {
		defer cancel()
		roomSession.SetParticipantLiveStatus(roleName, true)
		_, err = u.roomRepo.Update(newContext, roomSession)
		if err != nil {
			log.Println("Room session update error:", err)
		}
	}()

	err = u.hub.EnterRoom(c, roomCode, userCode)
	if err != nil {
		return err
	}

	go func() {
		defer cancel()

		if roleName == constant.RolesIsTerapis &&
			!roomSession.GetParticipantLiveStatus(constant.RoleIsUser) {
			if _, err := u.roomRepo.Delete(newContext, "socket_room_id", roomCode); err != nil {
				log.Println("Room session delete error:", err)
			}
			return
		}

		if roleName == constant.RoleIsUser &&
			!roomSession.GetParticipantLiveStatus(constant.RolesIsTerapis) {
			if _, err := u.roomRepo.Delete(newContext, "socket_room_id", roomCode); err != nil {
				log.Println("Room session delete error:", err)
			}
			return
		}

		roomSession.SetParticipantLiveStatus(roleName, false)
		if err := roomSession.SetParticipantSocketID(roleName, ""); err != nil {
			log.Println("Room session set error:", err)
			return
		}
		_, err = u.roomRepo.Update(newContext, roomSession)
		if err != nil {
			log.Println("Room session update error:", err)
		}
	}()

	return nil
}

func (u Usecase) CreateSession(ctx context.Context, roomCode string) (CreateSessionResponse, error) {
	roomSession, err := u.roomRepo.FindOne(ctx, "room_code", roomCode)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return CreateSessionResponse{}, err
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		login := u.Auth.GetUserLogin(ctx)
		if roomSession, err = u.newRoom(ctx, roomCode, login.UserId, login.RoleName); err != nil {
			return CreateSessionResponse{}, err
		}

		var sessionId string
		if id, err := roomSession.GetParticipantSocketID(login.RoleName); err != nil {
			return CreateSessionResponse{}, err
		} else if id != "" {
			sessionId = id
		}
		return CreateSessionResponse{
			SessionId: sessionId,
			RoomId:    roomSession.SocketRoomID,
		}, err

	}

	login := u.Auth.GetUserLogin(ctx)
	var sockerUserId string
	switch login.RoleName {
	case constant.RolesIsTerapis:
		if roomSession.EmployeeIsLive {
			return CreateSessionResponse{}, localerror.AccessControlError{Msg: "employee is already alive"}
		}
		roomSession.SocketEmployeeID = ""
	case constant.RoleIsUser:
		if roomSession.UserIsLive {
			return CreateSessionResponse{}, localerror.AccessControlError{Msg: "user is already alive"}
		}
		roomSession.SocketUserID = ""
	default:
		return CreateSessionResponse{}, localerror.AccessControlError{Msg: "invalid role"}
	}

	roomSession.SocketRoomID = ""
	_, err = u.roomRepo.Update(ctx, roomSession)
	if err != nil {
		return CreateSessionResponse{}, err

	}

	return CreateSessionResponse{
		SessionId: sockerUserId,
		RoomId:    roomSession.SocketRoomID,
	}, nil
}

func (u Usecase) newRoom(ctx context.Context, roomCode string, participantCode string, roleName string) (domain.RoomSession, error) {
	roomId, err := u.Davinci.GenerateUniqueKey(
		[]byte(u.Env.Get("SECRET_SOCKET_ROOM")),
		roomCode,
		10,
	)
	if err != nil {
		return domain.RoomSession{}, err
	}

	userId, err := u.Davinci.GenerateUniqueKeyWithPredicate(
		u.Env.Get("SECRET_SOCKET_USER"),
		participantCode,
		10,
		func(result string) (bool, error) {
			switch roleName {
			case constant.RoleIsUser:
				if exist, err := u.roomRepo.Exists(ctx, "socket_user_id", result); err != nil {
					return true, err
				} else if exist {
					return false, nil
				}
				return true, nil
			case constant.RolesIsTerapis:
				if exist, err := u.roomRepo.Exists(ctx, "socket_employee_id", result); err != nil {
					return true, err
				} else if exist {
					return false, nil
				}
				return true, nil
			}

			return false, localerror.AccessControlError{Msg: "invalid role"}
		},
	)
	if err != nil {
		return domain.RoomSession{}, err
	}

	roomSession := domain.RoomSession{
		SocketRoomID: roomId,
		RoomCode:     roomCode,
	}

	if err = roomSession.SetParticipantSocketID(roleName, userId); err != nil {
		return domain.RoomSession{}, err
	}

	_, err = u.roomRepo.Store(ctx, roomSession)
	if err != nil {
		return domain.RoomSession{}, err
	}
	return roomSession, nil
}
