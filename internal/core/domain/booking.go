package domain

import "time"

type Booking struct {
	BaseEntity
	Code        string    `json:"code"`
	UserID      uint      `gorm:"column:user_id;type:BIGINT UNSIGNED" json:"userId"`
	TherapistID uint      `gorm:"column:therapist_id;type:BIGINT UNSIGNED" json:"therapistId"`
	CityID      uint      `gorm:"column:city_id;type:BIGINT UNSIGNED" json:"cityId"`
	StatusID    uint      `gorm:"column:status_id;type:BIGINT UNSIGNED" json:"statusId"`
	DateTime    time.Time `gorm:"column:date_time" json:"dateTime"`
	Status      string    `gorm:"column:status" json:"status"`
	RefNumber   string    `gorm:"column:ref_number" json:"refNumber"`
	// Relations
	User       *UserExtended          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Therapist  *Therapist             `gorm:"foreignKey:TherapistID" json:"therapist,omitempty"`
	City       *MasterCity            `gorm:"foreignKey:CityID" json:"city,omitempty"`
	Payment    *Payment               `gorm:"foreignKey:BookingID" json:"payment,omitempty"`
	StatusHist []BookingStatusHistory `gorm:"foreignKey:BookingID" json:"statusHist,omitempty"`
}

func (b Booking) TableName() string {
	return "bookings"
}
