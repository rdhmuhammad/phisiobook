package user_management

import (
	"github.com/rdhmuhammad/phisiobook/shared/payload"
	"time"

	"iam_module/internal/core/domain"
)

type UserDetailItem struct {
	ID          uint       `json:"id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Role        string     `json:"role"`
	IsActive    string     `json:"isActive"`
	LastActive  *time.Time `json:"lastActive"`
	DestVisited int        `json:"destVisited"`
	JoinAt      time.Time  `json:"joinAt"`
}

type UserVisitedDestDetail struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	VisitedAt time.Time `json:"visitedAt"`
}

type UserVisitedDestQuery struct {
	UserID uint                          `bindQuery:"dataType=bigint" json:"userId" example:"1"`
	Filter *payload.GetListQueryNoPeriod `bindQuery:"dive=true" json:"filter"`
}

func DefaultUserDetailItem(item domain.UserEntityInterface) UserDetailItem {
	var status string
	if item.GetIsVerified() {
		status = domain.Active
	} else {
		status = domain.Inactive
	}

	return UserDetailItem{
		ID:         item.GetID(),
		Email:      item.GetEmail(),
		Name:       item.GetName(),
		JoinAt:     item.GetCreatedAt(),
		IsActive:   status,
		LastActive: item.GetLastActive(),
	}
}

type CreateUserRequest struct {
	ID        uint
	FullName  string `json:"fullName" example:"Jane Smith"`
	Email     string `json:"email" example:"jane@example.com"`
	Password  string `json:"password" example:"SecurePass123!"`
	RoleId    uint   `json:"roleId" example:"1"`
	StatusKey string `json:"statusKey" example:"active"`
}

// ===================== USER MOBILE ======================

type GetProfileResponse struct {
	ID             uint          `json:"id"`
	FullName       string        `json:"fullName"`
	Email          string        `json:"email"`
	LangActiveCode string        `json:"langActiveCode"`
	Lang           []ProfileLang `json:"lang"`
}

type ProfileLang struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
}

type UpdateUserLangRequest struct {
	Lang string `json:"lang" example:"en"`
}

type UserSubscription struct {
	Code      string     `json:"code"`
	IsActive  bool       `json:"isActive"`
	ExpiredAt *time.Time `json:"expiredAt"`
}

type NotifyNearestRequest struct {
	UserLongitude float64 `bindQuery:"dataType=float" json:"userLongitude" example:"106.8456"`
	UserLatitude  float64 `bindQuery:"dataType=float" json:"userLatitude" example:"-6.2088"`
}

type NearestDestItem struct {
	ID        uint    `gorm:"column:id" json:"id"`
	Name      string  `gorm:"column:title" json:"name"`
	Latitude  float64 `gorm:"column:latitude" json:"latitude"`
	Longitude float64 `gorm:"column:longitude" json:"longitude"`
}
type NotifyNearestResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Distance string `json:"distance"`
}
