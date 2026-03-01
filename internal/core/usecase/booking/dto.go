package booking

type CreateBookingRequest struct {
	CityID        uint   `json:"city_id"`
	DateTime      string `json:"dateTime"`
	TherapistCode string `json:"therapistCode"`
	PaymentMethod string `json:"payment_method"`
}

type UpdateStatus struct {
	Code   string
	Status string `json:"status"`
	Note   string `json:"note"`
}
