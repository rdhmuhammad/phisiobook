package domain

type SummaryHomepage struct {
	BaseEntity
	Terapis int `gorm:"column:terapis;type:int" json:"terapis"`
	Kota    int `gorm:"column:kota;type:int" json:"kota"`
	Rating  int `gorm:"column:rating;type:int" json:"rating"`
}

func (SummaryHomepage) TableName() string {
	return "summary_homepage"
}
