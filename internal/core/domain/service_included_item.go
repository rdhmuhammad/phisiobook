package domain

type ServiceIncludedItem struct {
	BaseEntity
	ServiceID uint   `gorm:"column:service_id;type:BIGINT UNSIGNED" json:"serviceId"`
	Name      string `gorm:"column:name;type:varchar(255)" json:"name"`

	// Relations
	Service *Service `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
}

func (s ServiceIncludedItem) TableName() string {
	return "service_included_items"
}
