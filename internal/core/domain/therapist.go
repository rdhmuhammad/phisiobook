package domain

import "database/sql"

type Therapist struct {
	BaseEntity
	Code           string            `json:"code"`
	Profile        string            `json:"profile"`
	Name           string            `json:"name"`
	IsVerified     int32             `json:"is_verified"`
	CityId         sql.Null[uint64]  `json:"city_id"`
	ExperienceYear int               `json:"experience_year"`
	Rating         float64           `json:"rating"`
	Price          int               `json:"price"`
	AuthID         uint              `json:"authId"`
	TherapyID      uint              `json:"therapyId"`
	TherapyType    MasterTherapyType `gorm:"foreignKey:TherapyID" json:"therapyType"`
}

func (receiver *Therapist) GetIsVerified() bool {
	return receiver.IsVerified == 1
}

func (receiver *Therapist) SetIsVerified(status bool) {
	switch status {
	case true:
		receiver.IsVerified = 1
		break
	case false:
		receiver.IsVerified = 0
		break
	default:
		receiver.IsVerified = 0
		break
	}
}

func (t Therapist) TableName() string {
	return "therapists"
}

func (t Therapist) GetProfile() string {
	return t.downloadFile(t.Profile)
}
