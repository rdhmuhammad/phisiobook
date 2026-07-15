package payload

import (
	"fmt"
	"time"
)

type GetListQueryNoPeriod struct {
	PerPage      int       `bindQuery:"dataType=integer" json:"perPage" example:"10"`
	Page         int       `bindQuery:"dataType=integer" json:"page" example:"1"`
	Search       string    `json:"search" example:"massage"`
	Date         time.Time `bindQuery:"dataType=timestamp" json:"date" example:"2024-01-15"`
	TimeZone     string    `json:"-"`
	ResponseType string
}

func (l *GetListQueryNoPeriod) SetIfEmpty() {
	if l.Page == 0 {
		l.Page = 1
	}

	if l.PerPage == 0 {
		l.PerPage = 15
	}

}

type GetListQuery struct {
	PerPage      int          `json:"perPage" example:"10"`
	Page         int          `json:"page" example:"1"`
	Search       string       `json:"search" example:"massage"`
	Date         time.Time    `json:"date" example:"2024-01-15"`
	FilterPeriod FilterPeriod `validate:"dive" json:"filterPeriod"`
	TimeZone     string       `json:"-"`
	ResponseType string
}

type Sorting struct {
	CreatedAt string `json:"createdAt" validate:"enum=ASC DESC"`
}

type FilterPeriod struct {
	Year       int       `json:"year" validate:"omitempty,gte=2022,lte=2100" example:"2024"`
	Month      string    `json:"month" validate:"omitempty,monthyearformat" example:"01-2024"`
	StartDate  time.Time `json:"startDate" example:"2024-01-01"`
	EndDate    time.Time `json:"endDate" example:"2024-01-31"`
	PeriodType string    `json:"periodType" example:"monthly"`
}

type GetListExpanseQuery struct {
	GetListQuery
	CategoryID int    `json:"categoryId" example:"1"`
	DanaSource string `json:"danaSource" example:"Operational"`
}

type GetListServiceQuery struct {
	GetListQuery
	UnitName string `json:"UnitName" example:"Jakarta Branch"`
}

type GetListTransactionQuery struct {
	GetListQuery
	Sorting
	CustomerID        uint   `json:"CustomerId" example:"1"`
	PaymentMethod     string `json:"paymentMethod" example:"bank_transfer"`
	PaymentStatus     string `json:"paymentStatus" example:"pending"`
	JenisTransaksi    string `json:"jenisTransaksi" example:"booking"`
	TransactionStatus string `json:"transactionStatus" example:"success"`
	ActionType        string `json:"actionType" example:"create"`
}

func (t GetListQuery) GetDateRange(tz *time.Location) (time.Time, time.Time) {
	var (
		month int
		year  int
	)

	_, err := fmt.Sscanf(t.FilterPeriod.Month, "%d-%d", &month, &year)
	if err != nil {
		return time.Time{}, time.Time{}
	}

	return time.Date(year, time.Month(month), 1, 0, 0, 0, 1, tz),
		time.Date(year, time.Month(month+1), -1, 23, 59, 59, 59, tz)

}
