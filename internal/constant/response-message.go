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
	InternalError

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

	// service-management
	CreateService
	UpdateService
	DeleteService
	GetDetailService
	GetListService
	GetCategories
	ServiceNotFound

	// chat
	GetChatList
	RoomChatNotFound
	RoomNotValid

	// booking
	BookingNotFound
	UpdateStatusBooking
	DropdownTherapistSuccess
	DropdownCitySuccess
	SettingIsNotSet

	// therapist
	RegisterTherapist
	TherapistEmailUsed

	// onboarding
	OnboardingSuccess
	TherapistNotFound

	// employee
	GetOnboardingList
	GetOnboardingDetail
	DeleteOnboarding
	OnboardingNotFound
)