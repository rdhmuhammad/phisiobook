package registration

type RegisterRequest struct {
	FullName string `json:"fullName" binding:"required" example:"John Doe"`
	Email    string `json:"email" binding:"required" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"SecurePass123!"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"SecurePass123!"`
	Timezone string `json:"timezone" example:"Asia/Jakarta"`
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
	Email   string `json:"email" example:"john@example.com"`
	Name    string `json:"name" example:"John Doe"`
	Content string `json:"content" example:"Your OTP code is 123456"`
	Subject string `json:"subject" example:"Verify Your Account"`
	UserID  uint64 `json:"userId" validate:"numeric" example:"1"`
}

type SendOtpResponse struct {
	Otp int32
}

type VerifyAccRequest struct {
	Email string `json:"email" example:"john@example.com"`
	Otp   int32  `json:"otp" example:"123456"`
}

type VerifyAccResponse struct {
	IsVerified bool `json:"isVerified"`
}
