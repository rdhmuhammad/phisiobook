package domain

import "time"

type UserEntityInterface interface {
	GetID() uint
	GetName() string
	GetEmail() string
	GetPassword() string
	GetIsVerified() bool
	GetAuthCode() string
	GetLastActive() *time.Time
	SetID(id uint)
	SetName(name string)
	SetEmail(email string)
	SetPassword(password string)
	SetIsVerified(status bool)
	SetAuthCode(code string)
	SetLastActive(t time.Time)
	GetCreatedAt() time.Time
	GetRoleName() string
	GetUpdatedAt() time.Time
}

type UserListItem struct {
	ID         uint       `gorm:"column:id" json:"id"`
	Email      string     `gorm:"column:email" json:"email"`
	Name       string     `gorm:"column:name" json:"name"`
	RoleKey    string     `gorm:"column:role_name" json:"roleKey"`
	StatusKey  string     `gorm:"column:status" json:"statusKey"`
	LastActive *time.Time `gorm:"column:last_active" json:"lastActive"`
}
