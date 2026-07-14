package homepage

import (
	"context"
	"fmt"

	"github.com/rdhmuhammad/phisiobook/internal/core/domain"
	"github.com/rdhmuhammad/phisiobook/pkg/db"
	"github.com/rdhmuhammad/phisiobook/shared/base"

	"gorm.io/gorm"
)

type Usecase struct {
	base.Port
	repo     db.GenericRepository[domain.SummaryHomepage]
	cityRepo db.GenericRepository[domain.MasterCity]
}

func New(dbConn *gorm.DB, prt base.Port) Usecase {
	return Usecase{
		Port:     prt,
		repo:     db.NewGenericeRepo(dbConn, domain.SummaryHomepage{}),
		cityRepo: db.NewGenericeRepo(dbConn, domain.MasterCity{}),
	}
}

func (uc Usecase) GetCityDropdown(ctx context.Context) ([]CityDropdownResponse, error) {
	cities, err := uc.cityRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]CityDropdownResponse, len(cities))
	for i, city := range cities {
		result[i] = CityDropdownResponse{
			ID:   city.ID,
			Name: city.Name,
		}
	}

	return result, nil
}

func (uc Usecase) GetSummaryHome(ctx context.Context) (*SummaryHomeResponse, error) {
	// Fetch the summary data from database
	summaries, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil, uc.ErrHandler.ErrorReturn(err)
	}

	// If no data exists, return default response
	if len(summaries) == 0 {
		return &SummaryHomeResponse{
			Terapis: 0,
			Kota:    "0",
			Rating:  0,
		}, nil
	}

	// Get the first (and typically only) summary record
	summary := summaries[0]

	return &SummaryHomeResponse{
		Terapis: summary.Terapis,
		Kota:    convertIntToString(summary.Kota),
		Rating:  float64(summary.Rating),
	}, nil
}

func convertIntToString(val int) string {
	return fmt.Sprintf("%d", val)
}
