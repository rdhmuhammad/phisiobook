package payload

import (
	"fmt"
	"time"
)

type GetListQueryNoPeriod struct {
	PerPage      int       `bindQuery:"dataType=integer" json:"perPage"`
	Page         int       `bindQuery:"dataType=integer" json:"page"`
	Search       string    `json:"search"`
	Date         time.Time `bindQuery:"dataType=timestamp" json:"date"`
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
	PerPage      int          `json:"perPage"`
	Page         int          `json:"page"`
	Search       string       `json:"search"`
	Date         time.Time    `json:"date"`
	FilterPeriod FilterPeriod `validate:"dive" json:"filterPeriod"`
	TimeZone     string       `json:"-"`
	ResponseType string
}

type Sorting struct {
	CreatedAt string `json:"createdAt" validate:"enum=ASC DESC"`
}

type FilterPeriod struct {
	Year       int       `json:"year" validate:"omitempty,gte=2022,lte=2100"`
	Month      string    `json:"month" validate:"omitempty,monthyearformat"`
	StartDate  time.Time `json:"startDate"`
	EndDate    time.Time `json:"endDate"`
	PeriodType string    `json:"periodType"`
}

type GetListExpanseQuery struct {
	GetListQuery
	CategoryID int    `json:"categoryId"`
	DanaSource string `json:"danaSource"`
}

type GetListServiceQuery struct {
	GetListQuery
	UnitName string `json:"UnitName"`
}

type GetListTransactionQuery struct {
	GetListQuery
	Sorting
	CustomerID        uint   `json:"CustomerId"`
	PaymentMethod     string `json:"paymentMethod"`
	PaymentStatus     string `json:"paymentStatus"`
	JenisTransaksi    string `json:"jenisTransaksi"`
	TransactionStatus string `json:"transactionStatus"`
	ActionType        string `json:"actionType"`
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
