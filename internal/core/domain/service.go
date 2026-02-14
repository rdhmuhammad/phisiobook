package domain

type Service struct {
	BaseEntity
	Name        string  `gorm:"column:name;type:varchar(255)" json:"name"`
	CategoryID  uint    `gorm:"column:category_id;type:BIGINT UNSIGNED" json:"categoryId"`
	Description string  `gorm:"column:description;type:text" json:"description"`
	Duration    int     `gorm:"column:duration;type:int" json:"duration"`
	Price       float64 `gorm:"column:price;type:decimal(15,2)" json:"price"`
	Commission  float64 `gorm:"column:commission;type:decimal(5,2)" json:"commission"`

	// Relations
	Category      *MasterServiceCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Areas         []ServiceArea          `gorm:"foreignKey:ServiceID" json:"areas,omitempty"`
	IncludedItems []ServiceIncludedItem  `gorm:"foreignKey:ServiceID" json:"includedItems,omitempty"`
}

func (s Service) TableName() string {
	return "services"
}
