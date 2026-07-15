package domain

type OnboardingApprovalHist struct {
	BaseEntity
	ApprovalID uint   `gorm:"column:approval_id" json:"approvalId"`
	Status     string `gorm:"column:status" json:"status"`
	Reason     string `gorm:"column:reason" json:"reason"`
}

func (o OnboardingApprovalHist) TableName() string {
	return "onboarding_approval_hist"
}
