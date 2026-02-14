package domain

type Setting struct {
	BaseEntity
	ApplicationFee float64 `gorm:"column:application_fee;type:decimal(10,2)" json:"applicationFee"`
	TaxPPN         float64 `gorm:"column:tax_ppn;type:decimal(5,2)" json:"taxPpn"`
	ServiceFee     float64 `gorm:"column:service_fee;type:decimal(10,2)" json:"serviceFee"`
}

func (s Setting) TableName() string {
	return "settings"
}
