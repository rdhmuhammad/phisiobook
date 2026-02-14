package domain

type MasterCity struct {
	BaseEntity
	Name string `json:"name"`
}

func (m MasterCity) TableName() string {
	return "master_cities"
}
