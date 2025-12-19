package domain

type MasterRole struct {
	BaseEntity
	Name  string `json:"name"`
	Label string `json:"label"`
}

func (m MasterRole) TableName() string {
	return "master_roles"
}
