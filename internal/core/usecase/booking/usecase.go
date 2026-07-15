package booking

import (
	"context"
	"errors"

	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/domain"
	"github.com/rdhmuhammad/phisiobook/pkg/db"
	"github.com/rdhmuhammad/phisiobook/pkg/localerror"
	"github.com/rdhmuhammad/phisiobook/pkg/mongodb"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	"github.com/rdhmuhammad/phisiobook/shared/payload"

	"strconv"

	"gorm.io/gorm"
)

type Usecase struct {
	base.Port
	dbConn         *gorm.DB
	cityRepo       db.GenericRepository[domain.MasterCity]
	statusHistRepo db.GenericRepository[domain.BookingStatusHistory]
	statusRepo     db.GenericRepository[domain.MasterBookingStatus]
	bookingRepo    db.GenericRepository[domain.Booking]
	chatRoomRepo   *mongodb.BaseRepo[domain.RoomSession]
	therapistRepo  db.GenericRepository[domain.Therapist]
	settingRepo    db.GenericRepository[domain.Setting]
	userRepo       db.GenericRepository[domain.UserExtended]
	userAdminRepo  db.GenericRepository[domain.UserAdminExtended]
}

func NewUsecase(dbCon *gorm.DB, mongoConn *mongodb.Conn, prt base.Port) Usecase {
	return Usecase{
		Port:           prt,
		dbConn:         dbCon,
		cityRepo:       db.NewGenericeRepo(dbCon, domain.MasterCity{}),
		statusHistRepo: db.NewGenericeRepo(dbCon, domain.BookingStatusHistory{}),
		statusRepo:     db.NewGenericeRepo(dbCon, domain.MasterBookingStatus{}),
		bookingRepo:    db.NewGenericeRepo(dbCon, domain.Booking{}),
		chatRoomRepo:   mongodb.NewBaseRepo(mongoConn, domain.RoomSession{}),
		therapistRepo:  db.NewGenericeRepo(dbCon, domain.Therapist{}),
		settingRepo:    db.NewGenericeRepo(dbCon, domain.Setting{}),
		userRepo:       db.NewGenericeRepo(dbCon, domain.UserExtended{}),
		userAdminRepo:  db.NewGenericeRepo(dbCon, domain.UserAdminExtended{}),
	}
}

func (uc Usecase) UpdateStatus(ctx context.Context, request UpdateStatus) (roomId string, err error) {
	booking, err := uc.bookingRepo.FindOneByExpression(ctx, db.Query(db.Equal(request.Code, "code")))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return roomId, localerror.InvalidData(constant.BookingNotFound.String())
		}
		return roomId, uc.ErrHandler.ErrorReturn(err)
	}

	trx := db.NewTransaction(uc.dbConn)
	defer func() {
		if err != nil {
			trx.End(err)
		} else {
			trx.End(nil)
		}
	}()

	status, err := uc.statusRepo.FindOneByExpression(ctx, db.Query(db.Equal(request.Status, "name")))
	if err != nil {
		return roomId, uc.ErrHandler.ErrorReturn(err)
	}
	userLogin := uc.Security.GetUserContext(ctx)

	history := domain.BookingStatusHistory{
		BookingID: booking.ID,
		StatusID:  status.ID,
		Notes:     request.Note,
	}
	history.SetCreated(userLogin.Email)
	_, err = db.GetRepo(trx, domain.BookingStatusHistory{}).
		Store(ctx, history)
	if err != nil {
		return roomId, uc.ErrHandler.ErrorReturn(err)
	}

	booking.SetUpdated(userLogin.Email)
	booking.Status = request.Status
	booking.StatusID = status.ID

	err = db.GetRepo(trx, domain.Booking{}).
		Update(ctx, booking)
	if err != nil {
		return roomId, uc.ErrHandler.ErrorReturn(err)
	}

	roomId = booking.RefNumber
	return roomId, nil
}

func (uc Usecase) CreateBooking(ctx context.Context, request CreateBookingRequest) error {
	var userLogin payload.SessionDataUser
	err := uc.Security.GetSessionLogin(ctx, &userLogin)
	if err != nil {
		return uc.ErrHandler.ErrorReturn(err)
	}

	pendingStatus, err := uc.statusRepo.FindOneByExpression(ctx, db.Query(db.Equal(constant.BookingStatusPending, "name")))
	if err != nil {
		return uc.ErrHandler.ErrorReturn(err)
	}

	trx := db.NewTransaction(uc.dbConn)
	defer func() {
		if err != nil {
			trx.End(err)
		} else {
			trx.End(nil)
		}
	}()

	type tg struct {
		ID     uint   `gorm:"column:id" json:"id"`
		UserID uint   `gorm:"column:auth_id" json:"userId"`
		Code   string `gorm:"column:code" json:"code"`
		Price  int    `gorm:"column:price" json:"price"`
	}
	var therapist = tg{}
	err = uc.therapistRepo.FindOneByExpSelection(ctx, &therapist, db.Query(db.Equal(request.TherapistCode, "code")))
	if err != nil {
		return uc.ErrHandler.ErrorReturn(err)
	}

	booking := domain.Booking{
		UserID:      userLogin.ID,
		TherapistID: therapist.ID,
		CityID:      request.CityID,
		StatusID:    pendingStatus.ID,
		DateTime:    uc.Clock.ParseWithTzFromCtx(ctx, request.DateTime, "2006-01-02 15:04:05"),
	}
	booking.SetCreated(userLogin.Email)

	bookingRepo := db.GetRepo(trx, domain.Booking{})
	booking, err = bookingRepo.Store(ctx, booking)
	if err != nil {
		return uc.ErrHandler.ErrorReturn(err)
	}

	refNumber, err := uc.Davinci.GenerateUniqueKeyWithPredicate(
		uc.Env.Get("BOOKING_SECRET"),
		strconv.Itoa(int(booking.ID)),
		6,
		func(result string) (bool, error) {
			exist, err := uc.bookingRepo.IsExist(ctx, "ref_number", result)
			if err != nil {
				return false, err
			}
			return exist, nil
		},
	)
	if err != nil {
		return uc.ErrHandler.ErrorReturn(err)
	}

	bookCode, err := uc.GenerateCode(ctx, "BOOK-", func(ctx context.Context, code string) (bool, error) {
		return uc.bookingRepo.IsExist(ctx, "code", code)
	})
	if err != nil {
		return uc.ErrHandler.ErrorReturn(err)
	}

	booking.RefNumber = refNumber
	booking.Code = bookCode
	err = bookingRepo.UpdateSelectedCols(ctx, booking, "ref_number", "code")
	if err != nil {
		return uc.ErrHandler.ErrorReturn(err)
	}

	statusHistory := domain.BookingStatusHistory{
		BookingID: booking.ID,
		StatusID:  pendingStatus.ID,
	}
	statusHistory.SetCreated(userLogin.UserReference)

	histRepo := db.GetRepo(trx, domain.BookingStatusHistory{})
	statusHistory, err = histRepo.Store(ctx, statusHistory)
	if err != nil {
		return uc.ErrHandler.ErrorReturn(err)
	}

	// Calculate amounts
	subTotal := therapist.Price
	var applicationFee float64
	var taxAmount float64
	totalAmount := float64(subTotal) + applicationFee + taxAmount

	if setting, err := uc.settingRepo.FindOneByExpression(ctx, db.Query(db.Equal(true, "status"))); err != nil {
		uc.ErrHandler.ErrorPrint(err)
	} else {
		applicationFee = setting.ApplicationFee
		taxAmount = (float64(subTotal) * setting.TaxPPN) / 100

	}

	// Create payment
	payment := domain.Payment{
		Total:           totalAmount,
		ReferenceNumber: refNumber,
		BookingID:       booking.ID,
		Method:          request.PaymentMethod,
		Status:          constant.PaymentStatusPending,
		ThirdPartyID:    0,
	}
	payment.SetCreated(userLogin.UserReference)

	paymentRepo := db.GetRepo(trx, domain.Payment{})
	payment, err = paymentRepo.Store(ctx, payment)
	if err != nil {
		return uc.ErrHandler.ErrorReturn(err)
	}

	// Create payment details
	paymentDetails := []domain.PaymentDetail{
		{
			Amount:          float64(subTotal),
			ReferenceNumber: refNumber,
			Name:            "sub_total",
			ParentPaymentID: payment.ID,
		},
		{
			Amount:          applicationFee,
			ReferenceNumber: refNumber,
			Name:            "application_fee",
			ParentPaymentID: payment.ID,
		},
		{
			Amount:          taxAmount,
			ReferenceNumber: refNumber,
			Name:            "tax_ppn",
			ParentPaymentID: payment.ID,
		},
	}

	paymentDetailRepo := db.GetRepo(trx, domain.PaymentDetail{})
	for _, detail := range paymentDetails {
		detail.SetCreated(userLogin.UserReference)
		_, err = paymentDetailRepo.Store(ctx, detail)
		if err != nil {
			return uc.ErrHandler.ErrorReturn(err)
		}
	}

	userRef, err := uc.Davinci.GenerateUniqueKeyWithPredicate(
		uc.Env.Get("CHAT_USER_SECRET"),
		strconv.Itoa(int(userLogin.ID)),
		6,
		func(result string) (bool, error) {
			return uc.userRepo.IsExist(ctx, "chat_ref", result)
		},
	)
	if err != nil {
		return uc.ErrHandler.ErrorReturn(err)
	}

	user := domain.UserExtended{
		ChatRef: userRef,
	}
	user.SetID(userLogin.ID)
	err = db.GetRepo(trx, domain.UserExtended{}).UpdateSelectedCols(ctx, user, "chat_ref")
	if err != nil {
		return uc.ErrHandler.ErrorReturn(err)
	}

	therapistRef, err := uc.Davinci.GenerateUniqueKeyWithPredicate(
		uc.Env.Get("CHAT_THERAPIST_SECRET"),
		strconv.Itoa(int(userLogin.ID)),
		6,
		func(result string) (bool, error) {
			return uc.userAdminRepo.IsExist(ctx, "chat_ref", result)
		},
	)
	if err != nil {
		return uc.ErrHandler.ErrorReturn(err)
	}

	therapistUser := domain.UserAdminExtended{}
	therapistUser.SetID(therapist.UserID)
	therapistUser.ChatRef = therapistRef
	err = db.GetRepo(trx, domain.UserAdminExtended{}).
		UpdateSelectedCols(ctx, therapistUser, "chat_ref")
	if err != nil {
		return uc.ErrHandler.ErrorReturn(err)
	}

	// TODO: add field chatId to therapist and customer
	_, err = uc.chatRoomRepo.Store(ctx, domain.RoomSession{
		BookCode:    refNumber,
		UserRef:     userRef,
		EmployeeRef: therapistRef,
		RoomIsLive:  false,
	})
	if err != nil {
		return uc.ErrHandler.ErrorReturn(err)
	}

	return nil
}

func (uc Usecase) GetAdjustPrice(ctx context.Context, therapistCode string) (AdjustPriceResponse, error) {
	type therapistSelection struct {
		Price int `gorm:"column:price" json:"price"`
	}

	var therapist therapistSelection
	err := uc.therapistRepo.FindOneByExpSelection(
		ctx,
		&therapist,
		db.Query(db.Equal(therapistCode, "code")),
	)
	if err != nil {
		return AdjustPriceResponse{}, uc.ErrHandler.ErrorReturn(err)
	}

	setting, err := uc.settingRepo.FindOneByExpression(ctx, db.Query(db.Equal(true, "status")))
	if err != nil {
		return AdjustPriceResponse{}, uc.ErrHandler.ErrorReturn(err)
	}

	return AdjustPriceResponse{
		SessionPrice: therapist.Price,
		AdminFee:     setting.ApplicationFee,
	}, nil
}

func (uc Usecase) RescheduleBooking(ctx context.Context, request RescheduleBookingRequest) (response RescheduleBookingResponse, err error) {
	userLogin := payload.SessionDataUser{}
	err = uc.Security.GetSessionLogin(ctx, &userLogin)
	if err != nil {
		return response, uc.ErrHandler.ErrorReturn(err)
	}

	booking, err := uc.bookingRepo.FindOneByExpression(ctx, db.Query(
		db.Equal(request.Code, "code"),
		db.Equal(userLogin.ID, "user_id"),
	))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response, localerror.InvalidData(constant.BookingNotFound.String())
		}
		return response, uc.ErrHandler.ErrorReturn(err)
	}

	activeStatuses, err := uc.statusRepo.FindAllByExpression(ctx, db.Query(
		db.InArray([]string{constant.BookingStatusPending, constant.BookingStatusScheduled}, "name"),
	))
	if err != nil {
		return response, uc.ErrHandler.ErrorReturn(err)
	}

	if !isActiveBooking(booking, activeStatuses) {
		return response, localerror.InvalidData(constant.BookingNotActive)
	}

	rescheduledAt := uc.Clock.ParseWithTzFromCtx(ctx, request.DateTime, "2006-01-02 15:04:05")
	if rescheduledAt.IsZero() || rescheduledAt.Before(uc.Clock.Now(ctx)) {
		return response, localerror.InvalidData(constant.InvalidBookingDateTime)
	}

	trx := db.NewTransaction(uc.dbConn)
	defer func() {
		if err != nil {
			trx.End(err)
		} else {
			trx.End(nil)
		}
	}()

	booking.DateTime = rescheduledAt
	booking.SetUpdated(userLogin.Email)
	err = db.GetRepo(trx, domain.Booking{}).
		UpdateSelectedCols(ctx, booking, "date_time", "updated_at", "updated_by")
	if err != nil {
		return response, uc.ErrHandler.ErrorReturn(err)
	}

	note := request.Note
	if note == "" {
		note = constant.RescheduleBookingHistNote
	}
	history := domain.BookingStatusHistory{
		BookingID: booking.ID,
		StatusID:  booking.StatusID,
		Notes:     note,
	}
	history.SetCreated(userLogin.Email)
	_, err = db.GetRepo(trx, domain.BookingStatusHistory{}).
		Store(ctx, history)
	if err != nil {
		return response, uc.ErrHandler.ErrorReturn(err)
	}

	return RescheduleBookingResponse{
		Code:      booking.Code,
		RefNumber: booking.RefNumber,
		DateTime:  rescheduledAt.Format("2006-01-02 15:04:05"),
		Status:    booking.Status,
	}, nil
}

func isActiveBooking(booking domain.Booking, statuses []domain.MasterBookingStatus) bool {
	for _, status := range statuses {
		if booking.StatusID == status.ID || booking.Status == status.Name {
			return true
		}
	}

	return booking.Status == constant.BookingStatusPending ||
		booking.Status == constant.BookingStatusScheduled
}
