package homepage

type SummaryHomeResponse struct {
	Terapis int     `json:"terapis"`
	Kota    string  `json:"kota"`
	Rating  float64 `json:"rating"`
}

type CityDropdownResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
