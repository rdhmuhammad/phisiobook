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

type AdjustPriceResponse struct {
	SessionPrice int     `json:"sessionPrice"`
	AdminFee     float64 `json:"adminFee"`
}

type RescheduleBookingRequest struct {
	Code     string `json:"-"`
	DateTime string `json:"dateTime" validate:"required"`
	Note     string `json:"note"`
}

type RescheduleBookingResponse struct {
	Code      string `json:"code"`
	RefNumber string `json:"refNumber"`
	DateTime  string `json:"dateTime"`
	Status    string `json:"status"`
}
