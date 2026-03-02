//go:generate stringer -type=ResponseMessage

package constant

type ResponseMessage int

const (
	LoginPasswordMismatch ResponseMessage = iota
	LoginUnverified
	RegisterEmailUsed
	EmailNotFound
	VerifyOtpExpired
	UserAlreadyVerified
	AccessNotAllowed
	SessionExpired

	// registration
	LogoutSuccess
	LoginSuccess
	RegisterSuccess
	VerifyOtpSuccess
	ResendOtpSuccess
	UserNotFound

	// user-management
	CreateUser
	UpdateUser
	DeleteUser
	GetDetailUser
	GetListUser
)
