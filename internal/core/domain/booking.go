package domain

import "time"

type Booking struct {
	BaseEntity
	UserID      uint      `gorm:"column:user_id;type:BIGINT UNSIGNED" json:"userId"`
	TherapistID uint      `gorm:"column:therapist_id;type:BIGINT UNSIGNED" json:"therapistId"`
	CityID      uint      `gorm:"column:city_id;type:BIGINT UNSIGNED" json:"cityId"`
	StatusID    uint      `gorm:"column:status_id;type:BIGINT UNSIGNED" json:"statusId"`
	DateTime    time.Time `gorm:"column:date_time" json:"dateTime"`

	// Relations
	User      *User                `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Therapist *Therapist           `gorm:"foreignKey:TherapistID" json:"therapist,omitempty"`
	City      *MasterCity          `gorm:"foreignKey:CityID" json:"city,omitempty"`
	Status    *MasterBookingStatus `gorm:"foreignKey:StatusID" json:"status,omitempty"`
}

func (b Booking) TableName() string {
	return "bookings"
}
