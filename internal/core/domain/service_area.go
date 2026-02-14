package domain

type ServiceArea struct {
	BaseEntity
	ServiceID uint `gorm:"column:service_id;type:BIGINT UNSIGNED" json:"serviceId"`
	CityID    uint `gorm:"column:city_id;type:BIGINT UNSIGNED" json:"cityId"`

	// Relations
	Service *Service    `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	City    *MasterCity `gorm:"foreignKey:CityID" json:"city,omitempty"`
}

func (s ServiceArea) TableName() string {
	return "service_areas"
}
