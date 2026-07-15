package booking

type CreateBookingRequest struct {
	CityID        uint   `json:"city_id" example:"1"`
	DateTime      string `json:"dateTime" example:"2024-01-15 10:00:00"`
	TherapistCode string `json:"therapistCode" example:"THR-ABC123"`
	PaymentMethod string `json:"payment_method" example:"bank_transfer"`
}

type UpdateStatus struct {
	Code   string
	Status string `json:"status" example:"completed"`
	Note   string `json:"note" example:"Booking has been completed"`
}

type AdjustPriceResponse struct {
	SessionPrice int     `json:"sessionPrice"`
	AdminFee     float64 `json:"adminFee"`
}

type RescheduleBookingRequest struct {
	Code     string `json:"-"`
	DateTime string `json:"dateTime" validate:"required" example:"2024-01-15 14:00:00"`
	Note     string `json:"note" example:"Rescheduled due to conflict"`
}

type RescheduleBookingResponse struct {
	Code      string `json:"code"`
	RefNumber string `json:"refNumber"`
	DateTime  string `json:"dateTime"`
	Status    string `json:"status"`
}
