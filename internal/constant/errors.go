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

	ErrQueryPeriod = fmt.Errorf("hanya diperbolehkan memilih salah satu filter")
)
