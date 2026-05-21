package history

import (
	"context"
	"strings"
	"time"

	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/domain"
	"github.com/rdhmuhammad/phisiobook/pkg/db"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	"github.com/rdhmuhammad/phisiobook/shared/payload"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Usecase struct {
	base.Port
	bookingRepo db.GenericRepository[domain.Booking]
}

func NewUsecase(dbConn *gorm.DB, prt base.Port) Usecase {
	return Usecase{
		Port:        prt,
		bookingRepo: db.NewGenericeRepo(dbConn, domain.Booking{}),
	}
}

func (uc Usecase) GetBookingHistory(ctx context.Context, request GetBookingHistoryRequest) (BookingHistoryResponse, error) {
	request.SetIfEmpty()

	userLogin, err := uc.getSession(ctx)
	if err != nil {
		return BookingHistoryResponse{}, err
	}

	now := uc.Clock.Now(ctx)
	summary, err := uc.getSummary(ctx, userLogin.ID, now)
	if err != nil {
		return BookingHistoryResponse{}, uc.ErrHandler.ErrorReturn(err)
	}

	expressions := uc.baseHistoryExpressions(userLogin.ID)
	expressions = uc.applyStatusFilter(expressions, request.Status, now)
	expressions = uc.applySearchFilter(expressions, request.Search)

	bookings, total, err := uc.bookingRepo.FindPagedByExpressionJoinOrdered(
		ctx,
		expressions,
		db.PaginationQuery{Page: request.Page, PerPage: request.PerPage},
		historyJoins(),
		historyPreloads(),
		db.ExpressionAnd,
		[]string{"bookings.date_time DESC", "bookings.id DESC"},
	)
	if err != nil {
		return BookingHistoryResponse{}, uc.ErrHandler.ErrorReturn(err)
	}

	items := make([]BookingHistoryItemResponse, len(bookings))
	for i, booking := range bookings {
		items[i] = mapHistoryBooking(booking, now)
	}

	pagination := BookingHistoryPagination{
		PerPage:     request.PerPage,
		Total:       total,
		CurrentPage: request.Page,
	}
	pagination.Evaluate()

	return BookingHistoryResponse{
		Summary:    summary,
		Histories:  items,
		Pagination: pagination,
	}, nil
}

func (uc Usecase) getSession(ctx context.Context) (userLogin payload.SessionDataUser, err error) {
	err = uc.Security.GetSessionLogin(ctx, &userLogin)
	if err != nil {
		return userLogin, uc.ErrHandler.ErrorReturn(err)
	}
	return userLogin, nil
}

func (uc Usecase) baseHistoryExpressions(userID uint) []clause.Expression {
	return db.Query(db.Equal(userID, "bookings.user_id"))
}

func (uc Usecase) getSummary(ctx context.Context, userID uint, now time.Time) (BookingHistorySummary, error) {
	baseExpressions := uc.baseHistoryExpressions(userID)

	total, err := uc.bookingRepo.CountByExpression(ctx, baseExpressions)
	if err != nil {
		return BookingHistorySummary{}, err
	}

	completed, err := uc.bookingRepo.CountByExpression(ctx,
		appendHistoryExpressions(baseExpressions, db.Equal(constant.BookingStatusCompleted, "bookings.status")))
	if err != nil {
		return BookingHistorySummary{}, err
	}

	upcoming, err := uc.bookingRepo.CountByExpression(ctx,
		appendHistoryExpressions(baseExpressions, upcomingStatusExpressions(now)...))
	if err != nil {
		return BookingHistorySummary{}, err
	}

	canceled, err := uc.bookingRepo.CountByExpression(ctx,
		appendHistoryExpressions(baseExpressions, db.Equal(constant.BookingStatusCanceled, "bookings.status")))
	if err != nil {
		return BookingHistorySummary{}, err
	}

	return BookingHistorySummary{
		TotalBooking: total,
		Completed:    completed,
		Upcoming:     upcoming,
		Canceled:     canceled,
	}, nil
}

func (uc Usecase) applyStatusFilter(expressions []clause.Expression, status string, now time.Time) []clause.Expression {
	switch NormalizeStatusFilter(status) {
	case StatusFilterUpcoming:
		return appendHistoryExpressions(expressions, upcomingStatusExpressions(now)...)
	case StatusFilterCompleted:
		return appendHistoryExpressions(expressions, db.Equal(constant.BookingStatusCompleted, "bookings.status"))
	case StatusFilterCanceled:
		return appendHistoryExpressions(expressions, db.Equal(constant.BookingStatusCanceled, "bookings.status"))
	default:
		return expressions
	}
}

func (uc Usecase) applySearchFilter(expressions []clause.Expression, search string) []clause.Expression {
	search = strings.TrimSpace(search)
	if search == "" {
		return expressions
	}

	return appendHistoryExpressions(expressions,
		db.Search(
			search,
			"bookings.code",
			"bookings.ref_number",
			"Therapist.name",
			"City.name",
			"Therapist__TherapyType.name",
		),
	)
}

func upcomingStatusExpressions(now time.Time) []clause.Expression {
	return db.Query(
		db.InArray([]string{constant.BookingStatusPending, constant.BookingStatusScheduled}, "bookings.status"),
		clause.Gte{Column: clause.Column{Table: "bookings", Name: "date_time"}, Value: now},
	)
}

func appendHistoryExpressions(base []clause.Expression, expressions ...clause.Expression) []clause.Expression {
	result := make([]clause.Expression, 0, len(base)+len(expressions))
	result = append(result, base...)
	result = append(result, expressions...)
	return result
}

func historyJoins() []string {
	return []string{"Therapist", "Therapist.TherapyType", "City"}
}

func historyPreloads() []string {
	return []string{"Therapist", "Therapist.TherapyType", "City", "Payment"}
}

func (p *BookingHistoryPagination) Evaluate() {
	if p.CurrentPage-1 > 0 {
		p.PreviousPage = p.CurrentPage - 1
	}

	if p.CurrentPage*p.PerPage < p.Total {
		p.NextPage = p.CurrentPage + 1
	}
}
