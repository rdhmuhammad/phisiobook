package domain

type MasterTherapyType struct {
	BaseEntity
	Name string `json:"name"`
}

func (m MasterTherapyType) TableName() string {
	return "master_therapy_types"
}
