package domain

type MasterServiceCategory struct {
	BaseEntity
	Name string `gorm:"column:name;type:varchar(100)" json:"name"`
}

func (m MasterServiceCategory) TableName() string {
	return "master_service_categories"
}
