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
	UserID uint `json:"userId"`
}

type LoginResponse struct {
	Email      string `json:"email"`
	UserID     uint   `json:"userId"`
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
