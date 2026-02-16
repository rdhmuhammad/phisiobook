package homepage

import (
	"base-be-golang/internal/core/domain"
	"base-be-golang/internal/core/port"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/db"
	"base-be-golang/pkg/logger"
	"base-be-golang/pkg/miniostorage"
	"context"
	"fmt"
	"gorm.io/gorm"
)

type Usecase struct {
	port.Port
	repo          db.GenericRepository[domain.SummaryHomepage]
	cityRepo      db.GenericRepository[domain.MasterCity]
	therapistRepo db.GenericRepository[domain.Therapist]
}

func New(dbConn *gorm.DB, dbCache cache.Cache, minioConn miniostorage.StorageMinio, rz *logger.ReZero) Usecase {
	return Usecase{
		therapistRepo: db.NewGenericeRepo(dbConn, domain.Therapist{}),
		Port:          port.NewPort(dbConn, dbCache, minioConn, rz),
		repo:          db.NewGenericeRepo(dbConn, domain.SummaryHomepage{}),
		cityRepo:      db.NewGenericeRepo(dbConn, domain.MasterCity{}),
	}
}

func (uc Usecase) GetSummaryHome(ctx context.Context) (*SummaryHomeResponse, error) {
	// Fetch the summary data from database
	summaries, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil, uc.Errhandler.ErrorReturn(err)
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

func (uc Usecase) GetTherapist(ctx context.Context, cityId uint) ([]TherapistDropdownResponse, error) {
	therapist, _, err := uc.therapistRepo.FindPagedByExpressionJoin(
		ctx,
		db.Query(db.Equal(cityId, "city_id")),
		db.PaginationQuery{Page: 1, PerPage: 5},
		[]string{"TherapyType"}, nil,
		db.ExpressionOr,
	)
	if err != nil {
		return nil, uc.Errhandler.ErrorReturn(err)
	}

	var result = make([]TherapistDropdownResponse, len(therapist))
	for i, t := range therapist {
		result[i] = TherapistDropdownResponse{
			Code:        t.Code,
			Name:        t.Name,
			Field:       t.TherapyType.Name,
			Profile:     t.GetProfile(),
			YearOnField: t.ExperienceYear,
			Price:       uc.FormatRupiah(t.Price),
			Rating:      float32(t.Rating),
		}
	}

	return result, nil
}
