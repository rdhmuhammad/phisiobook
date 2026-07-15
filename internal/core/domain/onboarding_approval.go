package domain

type OnboardingApproval struct {
	BaseEntity
	Code         string     `gorm:"column:code" json:"code"`
	TherapistID  uint       `gorm:"column:therapist_id" json:"therapistId"`
	Therapist    Therapist  `gorm:"foreignKey:TherapistID" json:"therapist"`
	LatestStatus string     `gorm:"column:latest_status" json:"latestStatus"`
	LatestReason string     `gorm:"column:latest_reason" json:"latestReason"`
	ApprovalByID *uint      `gorm:"column:approval_by_id" json:"approvalById"`
	ApprovalBy   *UserAdmin `gorm:"foreignKey:ApprovalByID" json:"approvalBy"`
}

func (o OnboardingApproval) TableName() string {
	return "onboarding_approvals"
}
