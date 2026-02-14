package domain

type Payment struct {
	BaseEntity
	Total           float64         `json:"total"`
	ReferenceNumber string          `json:"referenceNumber"`
	BookingID       uint            `json:"bookingId"`
	PaymentDate     []PaymentDetail `gorm:"foreignKey:ParentPaymentID" json:"paymentDate"`
	Method          string          `json:"method"`
	Status          string          `json:"status"`
	ThirdPartyID    uint            `json:"thirdPartyId"`
}

func (p Payment) TableName() string {
	return "payments"
}
