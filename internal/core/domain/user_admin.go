package domain

type UserAdmin struct {
	BaseEntity
	Code     string `json:"code"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	RoleID   uint   `gorm:"column:role_id" json:"roleID"`
	Password string `json:"password"`
	AuthCode string `gorm:"column:auth_code" json:"authCode"`
	IsActive int32  `gorm:"column:is_active" json:"isActive"`
	ChatRef  string `gorm:"column:chat_ref" json:"chatRef"`
}

func (u UserAdmin) TableName() string {
	return "user_admins"
}
