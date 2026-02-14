package domain

type Therapist struct {
	BaseEntity
	CityId         string  `json:"city_id"`
	Name           string  `json:"name"`
	ExperienceYear int     `json:"experience_year"`
	Rating         float64 `json:"rating"`
	Price          float64 `json:"price"`
	TherapyType    string  `json:"therapy_type"`
}

func (t Therapist) TableName() string {
	return "therapists"
}
