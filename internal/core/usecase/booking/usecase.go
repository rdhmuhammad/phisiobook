package booking

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/domain"
	"base-be-golang/internal/core/port"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/db"
	"base-be-golang/pkg/localerror"
	"base-be-golang/pkg/logger"
	"base-be-golang/pkg/miniostorage"
	"base-be-golang/pkg/mongodb"
	"context"
	"errors"
	"strconv"

	"gorm.io/gorm"
)

type Usecase struct {
	port.Port
	dbTrx          db.DBTransaction
	statusHistRepo db.GenericRepository[domain.BookingStatusHistory]
	statusRepo     db.GenericRepository[domain.MasterBookingStatus]
	bookingRepo    db.GenericRepository[domain.Booking]
	chatRoomRepo   mongodb.BaseRepo[domain.RoomSession]
	therapistRepo  db.GenericRepository[domain.Therapist]
	settingRepo    db.GenericRepository[domain.Setting]
	userRepo       db.GenericRepository[domain.User]
	userAdminRepo  db.GenericRepository[domain.UserAdmin]
}

func NewUsecase(dbCon *gorm.DB, dbCache cache.Cache, minioConn miniostorage.StorageMinio, rz *logger.ReZero) Usecase {
	return Usecase{
		statusHistRepo: db.NewGenericeRepo(dbCon, domain.BookingStatusHistory{}),
		userRepo:       db.NewGenericeRepo(dbCon, domain.User{}),
		userAdminRepo:  db.NewGenericeRepo(dbCon, domain.UserAdmin{}),
		bookingRepo:    db.NewGenericeRepo(dbCon, domain.Booking{}),
		Port:           port.NewPort(dbCon, dbCache, minioConn, rz),
		therapistRepo:  db.NewGenericeRepo(dbCon, domain.Therapist{}),
		settingRepo:    db.NewGenericeRepo(dbCon, domain.Setting{}),
		statusRepo:     db.NewGenericeRepo[domain.MasterBookingStatus](dbCon, domain.MasterBookingStatus{}),
	}
}

func (uc Usecase) UpdateStatus(ctx context.Context, request UpdateStatus) (roomId string, err error) {
	booking, err := uc.bookingRepo.FindOneByExpression(ctx, db.Query(db.Equal(request.Code, "code")))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return roomId, localerror.InvalidData(constant.BookingNotFound)
		}
		return roomId, uc.Errhandler.ErrorReturn(err)
	}

	uc.dbTrx.Begin()
	defer func(err error) {
		db.TransactionEnd(ctx, &uc.dbTrx, err)
	}(err)

	status, err := uc.statusRepo.FindOneByExpression(ctx, db.Query(db.Equal(request.Status, "name")))
	if err != nil {
		return roomId, uc.Errhandler.ErrorReturn(err)
	}
	userLogin := uc.Auth.GetUserLogin(ctx)

	history := domain.BookingStatusHistory{
		BookingID: booking.ID,
		StatusID:  status.ID,
		Notes:     request.Note,
	}
	history.SetCreated(userLogin.Email)
	_, err = db.GetRepo(&uc.dbTrx, domain.BookingStatusHistory{}).
		Store(ctx, history)
	if err != nil {
		return roomId, uc.Errhandler.ErrorReturn(err)
	}

	booking.SetUpdated(userLogin.Email)
	booking.Status = request.Status
	booking.StatusID = status.ID

	err = db.GetRepo(&uc.dbTrx, domain.Booking{}).
		Update(ctx, booking)
	if err != nil {
		return roomId, uc.Errhandler.ErrorReturn(err)
	}

	roomId = booking.RefNumber
	return roomId, nil
}

func (uc Usecase) CreateBooking(ctx context.Context, request CreateBookingRequest) error {
	userLogin, err := uc.GetUserLogin(ctx)
	if err != nil {
		return uc.Errhandler.ErrorReturn(err)
	}

	pendingStatus, err := uc.statusRepo.FindOneByExpression(ctx, db.Query(db.Equal(constant.BookingStatusPending, "name")))
	if err != nil {
		return uc.Errhandler.ErrorReturn(err)
	}

	uc.dbTrx.Begin()

	defer func(err error) {
		db.TransactionEnd(ctx, &uc.dbTrx, err)
	}(err)

	type tg struct {
		ID     uint   `gorm:"column:id" json:"id"`
		UserID uint   `gorm:"column:auth_id" json:"userId"`
		Code   string `gorm:"column:code" json:"code"`
		Price  int    `gorm:"column:price" json:"price"`
	}
	var therapist = tg{}
	err = uc.therapistRepo.FindOneByExpSelection(ctx, &therapist, db.Query(db.Equal(request.TherapistCode, "code")))
	if err != nil {
		return uc.Errhandler.ErrorReturn(err)
	}

	booking := domain.Booking{
		UserID:      userLogin.GetID(),
		TherapistID: therapist.ID,
		CityID:      request.CityID,
		StatusID:    pendingStatus.ID,
		DateTime:    uc.Clock.ParseWithTzFromCtx(ctx, request.DateTime, "2006-01-02 15:04:05"),
	}
	booking.SetCreated(userLogin.GetEmail())

	bookingRepo := db.GetRepo(&uc.dbTrx, domain.Booking{})
	booking, err = bookingRepo.Store(ctx, booking)
	if err != nil {
		return uc.Errhandler.ErrorReturn(err)
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
		return uc.Errhandler.ErrorReturn(err)
	}

	bookCode, err := uc.GenerateCode(ctx, "BOOK-", func(ctx context.Context, code string) (bool, error) {
		return uc.bookingRepo.IsExist(ctx, "code", code)
	})
	if err != nil {
		return uc.Errhandler.ErrorReturn(err)
	}

	booking.RefNumber = refNumber
	booking.Code = bookCode
	err = bookingRepo.UpdateSelectedCols(ctx, booking, "ref_number", "code")
	if err != nil {
		return uc.Errhandler.ErrorReturn(err)
	}

	statusHistory := domain.BookingStatusHistory{
		BookingID: booking.ID,
		StatusID:  pendingStatus.ID,
	}
	statusHistory.SetCreated(userLogin.GetAuthCode())

	histRepo := db.GetRepo(&uc.dbTrx, domain.BookingStatusHistory{})
	statusHistory, err = histRepo.Store(ctx, statusHistory)
	if err != nil {
		return uc.Errhandler.ErrorReturn(err)
	}

	setting, err := uc.settingRepo.FindOneByExpression(ctx, db.Query(db.Equal(true, "status")))
	if err != nil {
		return uc.Errhandler.ErrorReturn(err)
	}

	// Calculate amounts
	subTotal := therapist.Price
	applicationFee := setting.ApplicationFee
	taxAmount := (float64(subTotal) * setting.TaxPPN) / 100
	totalAmount := float64(subTotal) + applicationFee + taxAmount

	// Create payment
	payment := domain.Payment{
		Total:           totalAmount,
		ReferenceNumber: refNumber,
		BookingID:       booking.ID,
		Method:          request.PaymentMethod,
		Status:          constant.PaymentStatusPending,
		ThirdPartyID:    0,
	}
	payment.SetCreated(userLogin.GetAuthCode())

	paymentRepo := db.GetRepo(&uc.dbTrx, domain.Payment{})
	payment, err = paymentRepo.Store(ctx, payment)
	if err != nil {
		return uc.Errhandler.ErrorReturn(err)
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

	paymentDetailRepo := db.GetRepo(&uc.dbTrx, domain.PaymentDetail{})
	for _, detail := range paymentDetails {
		detail.SetCreated(userLogin.GetAuthCode())
		_, err = paymentDetailRepo.Store(ctx, detail)
		if err != nil {
			return uc.Errhandler.ErrorReturn(err)
		}
	}

	userRef, err := uc.Davinci.GenerateUniqueKeyWithPredicate(
		uc.Env.Get("CHAT_USER_SECRET"),
		strconv.Itoa(int(userLogin.GetID())),
		6,
		func(result string) (bool, error) {
			return uc.userRepo.IsExist(ctx, "chat_ref", result)
		},
	)
	if err != nil {
		return uc.Errhandler.ErrorReturn(err)
	}

	user := userLogin.(*domain.User)
	err = db.GetRepo(&uc.dbTrx, domain.User{}).UpdateSelectedCols(ctx, *user, "chat_ref")
	if err != nil {
		return uc.Errhandler.ErrorReturn(err)
	}

	therapistRef, err := uc.Davinci.GenerateUniqueKeyWithPredicate(
		uc.Env.Get("CHAT_THERAPIST_SECRET"),
		strconv.Itoa(int(userLogin.GetID())),
		6,
		func(result string) (bool, error) {
			return uc.userAdminRepo.IsExist(ctx, "chat_ref", result)
		},
	)
	if err != nil {
		return uc.Errhandler.ErrorReturn(err)
	}

	therapistUser := domain.UserAdmin{}
	therapistUser.SetID(therapist.UserID)
	therapistUser.ChatRef = therapistRef
	err = db.GetRepo(&uc.dbTrx, domain.UserAdmin{}).
		UpdateSelectedCols(ctx, therapistUser, "chat_ref")
	if err != nil {
		return uc.Errhandler.ErrorReturn(err)
	}

	// TODO: add field chatId to therapist and customer
	_, err = uc.chatRoomRepo.Store(ctx, domain.RoomSession{
		BookCode:    refNumber,
		UserRef:     userRef,
		EmployeeRef: therapistRef,
		RoomIsLive:  false,
	})
	if err != nil {
		return uc.Errhandler.ErrorReturn(err)
	}

	return nil
}
