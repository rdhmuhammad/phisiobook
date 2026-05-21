package history

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/domain"
)

const (
	StatusFilterAll       = "all"
	StatusFilterUpcoming  = "upcoming"
	StatusFilterCompleted = "completed"
	StatusFilterCanceled  = "canceled"
)

type GetBookingHistoryRequest struct {
	PerPage int    `bindQuery:"dataType=integer" json:"perPage"`
	Page    int    `bindQuery:"dataType=integer" json:"page"`
	Search  string `json:"search"`
	Status  string `bindQuery:"dataType=string" json:"status"`
}

func (r *GetBookingHistoryRequest) SetIfEmpty() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.PerPage == 0 {
		r.PerPage = 15
	}
	r.Status = NormalizeStatusFilter(r.Status)
}

type BookingHistoryResponse struct {
	Summary    BookingHistorySummary        `json:"summary"`
	Histories  []BookingHistoryItemResponse `json:"histories"`
	Pagination BookingHistoryPagination     `json:"pagination"`
}

type BookingHistorySummary struct {
	TotalBooking int `json:"totalBooking"`
	Completed    int `json:"completed"`
	Upcoming     int `json:"upcoming"`
	Canceled     int `json:"canceled"`
}

type BookingHistoryPagination struct {
	PerPage      int `json:"perPage"`
	Total        int `json:"total"`
	CurrentPage  int `json:"currentPage"`
	PreviousPage int `json:"previousPage"`
	NextPage     int `json:"nextPage"`
}

type BookingHistoryItemResponse struct {
	Code             string  `json:"code"`
	RefNumber        string  `json:"refNumber"`
	Status           string  `json:"status"`
	StatusLabel      string  `json:"statusLabel"`
	PaymentStatus    string  `json:"paymentStatus"`
	PaymentLabel     string  `json:"paymentLabel"`
	Total            float64 `json:"total"`
	TherapistCode    string  `json:"therapistCode"`
	TherapistName    string  `json:"therapistName"`
	TherapistProfile string  `json:"therapistProfile"`
	TherapyType      string  `json:"therapyType"`
	City             string  `json:"city"`
	DateTime         string  `json:"dateTime"`
	Date             string  `json:"date"`
	StartTime        string  `json:"startTime"`
	EndTime          string  `json:"endTime"`
	CanChat          bool    `json:"canChat"`
	CanReschedule    bool    `json:"canReschedule"`
	CanBookAgain     bool    `json:"canBookAgain"`
	CanInvoice       bool    `json:"canInvoice"`
}

func mapHistoryBooking(booking domain.Booking, now time.Time) BookingHistoryItemResponse {
	startAt := booking.DateTime
	endAt := startAt.Add(time.Hour)
	isUpcoming := isUpcomingStatus(booking.Status) && !startAt.Before(now)
	isCompleted := booking.Status == constant.BookingStatusCompleted
	isCanceled := booking.Status == constant.BookingStatusCanceled

	var therapistCode, therapistName, therapistProfile, therapyType string
	if booking.Therapist != nil {
		therapistCode = booking.Therapist.Code
		therapistName = booking.Therapist.Name
		therapistProfile = profileURL(booking.Therapist.Profile)
		therapyType = booking.Therapist.TherapyType.Name
	}

	var city string
	if booking.City != nil {
		city = booking.City.Name
	}

	var total float64
	var paymentStatus string
	if booking.Payment != nil {
		total = booking.Payment.Total
		paymentStatus = booking.Payment.Status
	}

	return BookingHistoryItemResponse{
		Code:             booking.Code,
		RefNumber:        booking.RefNumber,
		Status:           booking.Status,
		StatusLabel:      statusLabel(booking.Status),
		PaymentStatus:    paymentStatus,
		PaymentLabel:     paymentLabel(paymentStatus),
		Total:            total,
		TherapistCode:    therapistCode,
		TherapistName:    therapistName,
		TherapistProfile: therapistProfile,
		TherapyType:      therapyType,
		City:             city,
		DateTime:         startAt.Format("2006-01-02 15:04:05"),
		Date:             startAt.Format("2006-01-02"),
		StartTime:        startAt.Format("15:04"),
		EndTime:          endAt.Format("15:04"),
		CanChat:          isUpcoming,
		CanReschedule:    isUpcoming,
		CanBookAgain:     isCompleted || isCanceled,
		CanInvoice:       isCompleted,
	}
}

func profileURL(value string) string {
	if value == "" {
		return ""
	}

	return fmt.Sprintf("%s/api/v1/download?fileName=%s", os.Getenv("BACKEND_URL"), value)
}

func NormalizeStatusFilter(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "", "semua", StatusFilterAll:
		return StatusFilterAll
	case "akan_datang", "akan-datang", "upcoming", "active":
		return StatusFilterUpcoming
	case "selesai", "done", "completed", "complete":
		return StatusFilterCompleted
	case "dibatalkan", "cancelled", "canceled", "cancel":
		return StatusFilterCanceled
	default:
		return StatusFilterAll
	}
}

func isUpcomingStatus(status string) bool {
	return status == constant.BookingStatusPending || status == constant.BookingStatusScheduled
}

func statusLabel(status string) string {
	switch status {
	case constant.BookingStatusPending, constant.BookingStatusScheduled:
		return "Akan Datang"
	case constant.BookingStatusCompleted:
		return "Selesai"
	case constant.BookingStatusCanceled:
		return "Dibatalkan"
	default:
		return status
	}
}

func paymentLabel(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "paid", "success", "settlement", "capture", "lunas":
		return "Lunas"
	case constant.PaymentStatusPending, "":
		return "Menunggu Pembayaran"
	default:
		return status
	}
}
