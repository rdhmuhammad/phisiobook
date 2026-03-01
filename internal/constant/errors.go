package constant

import "fmt"

var (

	// auth
	LoginSuccess     = "LoginSuccess"
	RegisterSuccess  = "RegisterSuccess"
	VerifyOtpSuccess = "VerifyOtpSuccess"
	ResendOtpSuccess = "ResendOtpSuccess"
	LogoutSuccess    = "LogoutSuccess"

	AccessNotAllowed      = "AccessNotAllowed"
	SessionExpired        = "SessionExpired"
	LoginPasswordMismatch = "LoginPasswordMismatch"
	LoginUnverified       = "LoginUnverified"
	RegisterEmailUsed     = "RegisterEmailUsed"
	EmailNotFound         = "EmailNotFound"
	VerifyOtpExpired      = "VerifyOtpExpired"
	UserAlreadyVerified   = "UserAlreadyVerified"

	InternalError  = "InternalError"
	ErrQueryPeriod = fmt.Errorf("hanya diperbolehkan memilih salah satu filter")

	// User Management
	CreateUser    = "CreateUserSuccess"
	UpdateUser    = "UpdateUserSuccess"
	DeleteUser    = "DeleteUserSuccess"
	GetDetailUser = "GetDetailUserSuccess"
	GetListUser   = "GetListUserSuccess"

	// Service Management
	CreateService    = "CreateServiceSuccess"
	UpdateService    = "UpdateServiceSuccess"
	DeleteService    = "DeleteServiceSuccess"
	GetDetailService = "GetDetailServiceSuccess"
	GetListService   = "GetListServiceSuccess"
	GetCategories    = "GetCategoriesSuccess"

	// Not Found
	UserNotFound     = "UserNotFound"
	ServiceNotFound  = "ServiceNotFound"
	RoomChatNotFound = "RoomChatNotFound"
	RoomNotValid     = "RoomNotValid"
	BookingNotFound  = "BookingNotFound"

	// Booking
	UpdateStatusBooking = "UpdateStatusBooking"
)
