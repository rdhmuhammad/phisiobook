package homepage

import (
	"base-be-golang/internal/core/domain"
	"base-be-golang/internal/core/port"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/db"
	"base-be-golang/pkg/miniostorage"
	"context"
	"fmt"
	"gorm.io/gorm"
)

type Usecase struct {
	port.Port
	repo db.GenericRepository[domain.SummaryHomepage]
}

func New(dbConn *gorm.DB, dbCache cache.Cache, minioConn miniostorage.StorageMinio) Usecase {
	return Usecase{
		Port: port.NewPort(dbConn, dbCache, minioConn),
		repo: db.NewGenericeRepo(dbConn, domain.SummaryHomepage{}),
	}
}

func (uc Usecase) GetSummaryHome(ctx context.Context) (*SummaryHomeResponse, error) {
	// Fetch the summary data from database
	summaries, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil, err
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
