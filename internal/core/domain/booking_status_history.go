package domain

type BookingStatusHistory struct {
	BaseEntity
	BookingID uint   `gorm:"column:booking_id;type:BIGINT UNSIGNED" json:"bookingId"`
	StatusID  uint   `gorm:"column:status_id;type:BIGINT UNSIGNED" json:"statusId"`
	Notes     string `gorm:"column:notes;type:text" json:"notes"`

	// Relations
	Booking *Booking             `gorm:"foreignKey:BookingID" json:"booking,omitempty"`
	Status  *MasterBookingStatus `gorm:"foreignKey:StatusID" json:"status,omitempty"`
}

func (b BookingStatusHistory) TableName() string {
	return "booking_status_histories"
}
