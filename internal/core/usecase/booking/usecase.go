package booking

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/domain"
	"base-be-golang/internal/core/port"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/db"
	"base-be-golang/pkg/miniostorage"
	"context"
	"fmt"
	"gorm.io/gorm"
	"log"
)

type Usecase struct {
	port.Port
	dbTrx         db.DBTransaction
	statusRepo    db.GenericRepository[domain.MasterBookingStatus]
	therapistRepo db.GenericRepository[domain.Therapist]
	settingRepo   db.GenericRepository[domain.Setting]
}

func NewUsecase(dbCon *gorm.DB, dbCache cache.Cache, minioConn miniostorage.StorageMinio) Usecase {
	return Usecase{
		Port:          port.NewPort(dbCon, dbCache, minioConn),
		therapistRepo: db.NewGenericeRepo(dbCon, domain.Therapist{}),
		settingRepo:   db.NewGenericeRepo(dbCon, domain.Setting{}),
		statusRepo:    db.NewGenericeRepo[domain.MasterBookingStatus](dbCon, domain.MasterBookingStatus{}),
	}
}

func (uc Usecase) CreateBooking(ctx context.Context, request CreateBookingRequest) error {
	userLogin, err := uc.GetUserLogin(ctx)
	if err != nil {
		return err
	}

	pendingStatus, err := uc.statusRepo.FindOneByExpression(ctx, db.Query(db.Equal(constant.BookingStatusPending, "name")))
	if err != nil {
		return err
	}

	uc.dbTrx.Begin()
	defer func(err error) {
		if r := recover(); r != nil {
			err := uc.dbTrx.End(fmt.Errorf("recovery"))
			if err != nil {
				log.Println(err)
				return
			}
		}

		err = uc.dbTrx.End(err)
		if err != nil {
			log.Println(err)
			return
		}
	}(err)

	booking := domain.Booking{
		UserID:      userLogin.GetID(),
		TherapistID: request.TherapistID,
		CityID:      request.CityID,
		StatusID:    pendingStatus.ID,
		DateTime:    uc.Clock.ParseWithTzFromCtx(ctx, request.DateTime, "2006-01-02 15:04:05"),
	}
	booking.SetCreated(userLogin.GetEmail())

	bookingRepo := db.GetRepo(&uc.dbTrx, domain.Booking{})
	booking, err = bookingRepo.Store(ctx, booking)
	if err != nil {
		return err
	}

	statusHistory := domain.BookingStatusHistory{
		BookingID: booking.ID,
		StatusID:  pendingStatus.ID,
	}
	statusHistory.SetCreated(userLogin.GetAuthCode())

	histRepo := db.GetRepo(&uc.dbTrx, domain.BookingStatusHistory{})
	statusHistory, err = histRepo.Store(ctx, statusHistory)
	if err != nil {
		return err
	}

	therapist, err := uc.therapistRepo.FindOneByID(ctx, request.TherapistID)
	if err != nil {
		return err
	}

	setting, err := uc.settingRepo.FindOneByExpression(ctx, db.Query(db.Equal(true, "status")))
	if err != nil {
		return err
	}

	// Calculate amounts
	subTotal := therapist.Price
	applicationFee := setting.ApplicationFee
	taxAmount := (subTotal * setting.TaxPPN) / 100
	totalAmount := subTotal + applicationFee + taxAmount

	// Generate reference number
	referenceNumber := fmt.Sprintf("BK-%d-%d", booking.ID, uc.Clock.NowUnix())

	// Create payment
	payment := domain.Payment{
		Total:           totalAmount,
		ReferenceNumber: referenceNumber,
		BookingID:       booking.ID,
		Method:          request.PaymentMethod,
		Status:          "pending",
		ThirdPartyID:    0,
	}
	payment.SetCreated(userLogin.GetAuthCode())

	paymentRepo := db.GetRepo(&uc.dbTrx, domain.Payment{})
	payment, err = paymentRepo.Store(ctx, payment)
	if err != nil {
		return err
	}

	// Create payment details
	paymentDetails := []domain.PaymentDetail{
		{
			Amount:          subTotal,
			ReferenceNumber: referenceNumber,
			Name:            "sub_total",
			ParentPaymentID: payment.ID,
		},
		{
			Amount:          applicationFee,
			ReferenceNumber: referenceNumber,
			Name:            "application_fee",
			ParentPaymentID: payment.ID,
		},
		{
			Amount:          taxAmount,
			ReferenceNumber: referenceNumber,
			Name:            "tax_ppn",
			ParentPaymentID: payment.ID,
		},
	}

	paymentDetailRepo := db.GetRepo(&uc.dbTrx, domain.PaymentDetail{})
	for _, detail := range paymentDetails {
		detail.SetCreated(userLogin.GetAuthCode())
		_, err = paymentDetailRepo.Store(ctx, detail)
		if err != nil {
			return err
		}
	}

	return nil
}
