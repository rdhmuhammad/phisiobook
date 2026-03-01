package domain

import (
	"base-be-golang/internal/constant"
	"base-be-golang/pkg/localerror"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoomSession struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	SocketUserID     string             `json:"socket_user_id" bson:"socket_user_id"`
	UserFullName     string             `json:"user_full_name" bson:"user_full_name"`
	UserRef          string             `json:"user_ref" bson:"user_ref"`
	UserIsLive       bool               `json:"user_is_live" bson:"user_is_live"`
	SocketEmployeeID string             `json:"socket_employee_id" bson:"socket_employee_id"`
	EmployeeFullName string             `json:"employee_full_name" bson:"employee_full_name"`
	EmployeeRef      string             `json:"employee_ref" bson:"employee_ref"`
	EmployeeIsLive   bool               `json:"employee_is_live" bson:"employee_is_live"`
	BookCode         string             `json:"book_code" bson:"book_code"`
	IsValid          bool               `json:"is_valid" bson:"is_valid"`
	RoomIsLive       bool               `json:"room_is_live" bson:"room_is_live"`
}

func (receiver RoomSession) ParticipantSocketIDEqual(roleName string, id string) bool {
	switch roleName {
	case constant.RoleIsUser:
		return receiver.SocketUserID == id
	case constant.RolesIsTerapis:
		return receiver.SocketEmployeeID == id
	}
	return false
}

func (receiver *RoomSession) GetParticipantLiveStatus(role string) bool {
	switch role {
	case constant.RolesIsTerapis:
		return receiver.EmployeeIsLive
	case constant.RoleIsUser:
		return receiver.UserIsLive
	}

	return false
}

func (receiver *RoomSession) SetParticipantLiveStatus(role string, status bool) {
	switch role {
	case constant.RolesIsTerapis:
		receiver.EmployeeIsLive = status
	case constant.RoleIsUser:
		receiver.UserIsLive = status
	}
}

func (receiver *RoomSession) GetParticipantSocketID(role string) (string, error) {
	switch role {
	case constant.RolesIsTerapis:
		return receiver.SocketEmployeeID, nil
	case constant.RoleIsUser:
		return receiver.SocketUserID, nil
	default:
		return "", localerror.AccessControlError{Msg: "invalid role"}
	}
}

func (receiver *RoomSession) GetActorName(userRef string) string {
	switch userRef {
	case receiver.UserRef:
		return receiver.UserFullName
	case receiver.EmployeeRef:
		return receiver.EmployeeFullName
	}
	return ""
}

func (receiver *RoomSession) GetToSocketID(userRef string) string {
	switch userRef {
	case receiver.UserRef:
		return receiver.EmployeeRef
	case receiver.EmployeeRef:
		return receiver.EmployeeRef
	}

	return ""
}

func (receiver *RoomSession) SetParticipantSocketID(role string, id string) error {
	switch role {
	case constant.RolesIsTerapis:
		receiver.SocketEmployeeID = id
		return nil
	case constant.RoleIsUser:
		receiver.SocketUserID = id
		return nil
	default:
		return localerror.AccessControlError{Msg: "invalid role"}
	}
}

func (receiver RoomSession) GetID() primitive.ObjectID {
	return receiver.ID
}

func (receiver RoomSession) GetCollectionName() string {
	return "room_sessions"
}
