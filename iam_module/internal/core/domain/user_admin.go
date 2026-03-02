package domain

import (
	"database/sql"
	"time"
)

type UserAdmin struct {
	BaseEntity
	Code       string       `json:"code"`
	FullName   string       `json:"fullName"`
	Email      string       `json:"email"`
	Phone      string       `json:"phone"`
	Role       MasterRole   `gorm:"foreignKey:RoleID" json:"role"`
	RoleID     uint         `gorm:"column:role_id" json:"roleID"`
	Password   string       `json:"password"`
	AuthCode   string       `json:"authCode"`
	IsActive   int32        `gorm:"column:is_active" json:"isActive"`
	LastActive sql.NullTime `gorm:"column:last_active" json:"lastActive"`
}

const (
	Active   = "active"
	Inactive = "inactive"
)

func (receiver *UserAdmin) GetRoleName() string {
	return receiver.Role.Name
}

func (receiver *UserAdmin) SetLastActive(t time.Time) {
	receiver.LastActive = sql.NullTime{Time: t, Valid: true}
}
func (u *UserAdmin) GetLastActive() *time.Time {
	if u.LastActive.Valid {
		return &u.LastActive.Time
	} else {
		return nil
	}
}

func (receiver *UserAdmin) GetIsVerified() bool {
	return receiver.IsActive == 1
}

func (receiver *UserAdmin) GetStatusKey() string {
	switch receiver.GetIsVerified() {
	case true:
		return Active
	case false:
		return Inactive
	}
	return ""
}

func (receiver *UserAdmin) SetIsVerified(status bool) {
	switch status {
	case true:
		receiver.IsActive = 1
		break
	case false:
		receiver.IsActive = 0
		break
	default:
		receiver.IsActive = 0
		break
	}
}

func (receiver *UserAdmin) GetID() uint                 { return receiver.ID }
func (receiver *UserAdmin) GetName() string             { return receiver.FullName }
func (receiver *UserAdmin) GetEmail() string            { return receiver.Email }
func (receiver *UserAdmin) GetPassword() string         { return receiver.Password }
func (receiver *UserAdmin) GetAuthCode() string         { return receiver.AuthCode }
func (receiver *UserAdmin) SetID(id uint)               { receiver.ID = id }
func (receiver *UserAdmin) SetName(name string)         { receiver.FullName = name }
func (receiver *UserAdmin) SetEmail(email string)       { receiver.Email = email }
func (receiver *UserAdmin) SetPassword(password string) { receiver.Password = password }
func (receiver *UserAdmin) SetAuthCode(code string)     { receiver.AuthCode = code }
func (receiver UserAdmin) TableName() string {
	return "user_admins"
}
