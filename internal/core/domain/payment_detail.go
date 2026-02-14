package domain

type PaymentDetail struct {
	BaseEntity
	Amount          float64 `json:"amount"`
	ReferenceNumber string  `json:"referenceNumber"`
	Name            string  `json:"name"`
	ParentPaymentID uint    `json:"parentPaymentId"`
}

func (p PaymentDetail) TableName() string {
	return "payment_details"
}
