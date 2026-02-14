package booking

type CreateBookingRequest struct {
	CityID        uint   `json:"city_id"`
	DateTime      string `json:"dateTime"`
	TherapistID   uint   `json:"therapistId"`
	PaymentMethod string `json:"payment_method"`
}
