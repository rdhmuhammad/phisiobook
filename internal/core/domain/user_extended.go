package domain

type UserExtended struct {
	BaseEntity
	ChatRef string `json:"chatRef"`
}

func (e UserExtended) TableName() string {
	return "users"
}

type UserAdminExtended struct {
	BaseEntity
	ChatRef string `json:"chatRef"`
}

func (u UserAdminExtended) TableName() string {
	return "user_admins"
}
