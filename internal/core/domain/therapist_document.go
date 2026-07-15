package domain

type TherapistDocument struct {
	BaseEntity
	TherapistID uint   `gorm:"column:therapist_id" json:"therapistId"`
	KtpDoc      string `gorm:"column:ktp_doc" json:"ktpDoc"`
	StrDoc      string `gorm:"column:str_doc" json:"strDoc"`
	SipDoc      string `gorm:"column:sip_doc" json:"sipDoc"`
	IjazahDoc   string `gorm:"column:ijazah_doc" json:"ijazahDoc"`
	BankName    string `gorm:"column:bank_name" json:"bankName"`
	BankCode    string `gorm:"column:bank_code" json:"bankCode"`
	AccName     string `gorm:"column:acc_name" json:"accName"`
	AccNumber   string `gorm:"column:acc_number" json:"accNumber"`
}

func (t TherapistDocument) TableName() string {
	return "therapist_documents"
}

