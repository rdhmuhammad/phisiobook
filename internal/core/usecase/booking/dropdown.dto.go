package booking

type CityDropdownResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type TherapistDropdownResponse struct {
	Code        string  `json:"code"`
	Price       string  `json:"price"`
	Name        string  `json:"name"`
	Field       string  `json:"field"`
	Profile     string  `json:"profile"`
	YearOnField int     `json:"yearOnField"`
	Rating      float32 `json:"rating"`
}
