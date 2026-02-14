package domain

type MasterBookingStatus struct {
	BaseEntity
	Name string `gorm:"column:name;type:varchar(50)" json:"name"`
	Code string `gorm:"column:code;type:varchar(50)" json:"code"`
}

func (m MasterBookingStatus) TableName() string {
	return "master_booking_statuses"
}
