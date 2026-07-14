package registration

type RegisterRequest struct {
	FullName string `json:"fullName" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Timezone string `json:"timezone"`
	Role     string `json:"-"`
}

type RegisterResponse struct {
	Code string `json:"code"`
}

type LoginResponse struct {
	Email      string `json:"email"`
	Lang       string
	Code       string `json:"code"`
	Token      string `json:"token"`
	IsVerified bool   `json:"isVerified"`
}

type SendOtpRequest struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Subject string `json:"subject"`
	UserID  uint64 `json:"userId" validate:"numeric"`
}

type SendOtpResponse struct {
	Otp int32
}

type VerifyAccRequest struct {
	Email string `json:"email"`
	Otp   int32  `json:"otp"`
}

type VerifyAccResponse struct {
	IsVerified bool `json:"isVerified"`
}
