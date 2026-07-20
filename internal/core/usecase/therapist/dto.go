package therapist

import (
	"io"
	"time"

	"github.com/rdhmuhammad/phisiobook/shared/payload"
)

type RegisterTherapistRequest struct {
	FullName string `form:"fullName" binding:"required" example:"Dr. Sarah Johnson"`
	Email    string `form:"email"    binding:"required" example:"sarah@example.com"`
	Password string `form:"password" binding:"required" example:"SecurePass123!"`
	Phone    string `form:"phone"    binding:"required" example:"081234567890"`
	Profile  FileInfo
}

type RegisterTherapistResponse struct {
	Code  string `json:"code"`
	Email string `json:"email"`
}

type FileInfo struct {
	Reader    io.Reader
	Size      int64
	Extension string
}

type OnboardingRequest struct {
	KtpFile    FileInfo
	SipFile    FileInfo
	StrFile    FileInfo
	IjazahFile FileInfo
	BankCode   string
	AccName    string
	AccNumber  string
}

type OnboardingResponse struct {
	Code   string `json:"code"`
	Status string `json:"status"`
}

type OnboardingListQuery struct {
	PerPage      int    `bindQuery:"dataType=integer" json:"perPage" example:"10"`
	Page         int    `bindQuery:"dataType=integer" json:"page" example:"1"`
	Search       string `json:"search" example:"john"`
	LatestStatus string `json:"latestStatus" example:"PENDING"`
}

func (l *OnboardingListQuery) SetIfEmpty() {
	if l.Page == 0 {
		l.Page = 1
	}
	if l.PerPage == 0 {
		l.PerPage = 15
	}
}

type OnboardingListItem struct {
	Code           string    `json:"code"`
	TherapistName  string    `json:"therapistName"`
	LatestStatus   string    `json:"latestStatus"`
	LatestReason   string    `json:"latestReason"`
	CreatedAt      time.Time `json:"createdAt"`
	ApprovalByName string    `json:"approvalByName"`
}

type OnboardingListResponse = payload.PaginationResponse[OnboardingListItem]

type OnboardingDetailResponse struct {
	Code             string    `json:"code"`
	TherapistCode    string    `json:"therapistCode"`
	TherapistName    string    `json:"therapistName"`
	TherapistProfile string    `json:"therapistProfile"`
	LatestStatus     string    `json:"latestStatus"`
	LatestReason     string    `json:"latestReason"`
	ApprovalByName   string    `json:"approvalByName"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	KtpDoc           string    `json:"ktpDoc"`
	StrDoc           string    `json:"strDoc"`
	SipDoc           string    `json:"sipDoc"`
	IjazahDoc        string    `json:"ijazahDoc"`
	BankName         string    `json:"bankName"`
	BankCode         string    `json:"bankCode"`
	AccName          string    `json:"accName"`
	AccNumber        string    `json:"accNumber"`
}
