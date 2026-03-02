package domain

import (
	"database/sql"
	"time"
)

type User struct {
	BaseEntity
	Code       string       `json:"code"`
	Profile    string       `json:"profile"`
	FullName   string       `json:"fullName"`
	Phone      string       `json:"phone"`
	Email      string       `json:"email"`
	Password   string       `json:"password"`
	IsVerified int32        `gorm:"column:is_verified" json:"isVerified"`
	OTPCode    int32        `gorm:"column:otp_code" json:"otpCode"`
	AuthCode   string       `json:"authCode"`
	Lang       string       `json:"lang"`
	LastActive sql.NullTime `gorm:"column:last_active" json:"lastActive"`
}

func (receiver User) GetRoleName() string {
	return ""
}

func (receiver *User) GetIsVerified() bool {
	return receiver.IsVerified == 1
}

func (receiver *User) SetIsVerified(status bool) {
	switch status {
	case true:
		receiver.IsVerified = 1
		break
	case false:
		receiver.IsVerified = 0
		break
	default:
		receiver.IsVerified = 0
		break
	}
}

func (u *User) GetLastActive() *time.Time {
	if u.LastActive.Valid {
		return &u.LastActive.Time
	} else {
		return nil
	}
}
func (receiver *User) SetLastActive(t time.Time) {
	receiver.LastActive = sql.NullTime{Time: t, Valid: true}
}
func (receiver *User) GetID() uint                 { return receiver.ID }
func (receiver *User) GetName() string             { return receiver.FullName }
func (receiver *User) GetEmail() string            { return receiver.Email }
func (receiver *User) GetPassword() string         { return receiver.Password }
func (receiver *User) GetAuthCode() string         { return receiver.AuthCode }
func (receiver *User) SetID(id uint)               { receiver.ID = id }
func (receiver *User) SetName(name string)         { receiver.FullName = name }
func (receiver *User) SetEmail(email string)       { receiver.Email = email }
func (receiver *User) SetPassword(password string) { receiver.Password = password }
func (receiver *User) SetAuthCode(code string)     { receiver.AuthCode = code }
func (receiver User) TableName() string {
	return "users"
}
