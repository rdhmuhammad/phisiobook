package health

import (
	"context"
	"github.com/rdhmuhammad/phisiobook/shared/base"

	"gorm.io/gorm"
)

type Usecase struct {
	base.Port
	dbGorm *gorm.DB
}

func New(dbConn *gorm.DB, prt base.Port) Usecase {
	return Usecase{
		Port:   prt,
		dbGorm: dbConn,
	}
}

func (uc Usecase) CheckHealth(ctx context.Context) (map[string]string, error) {
	status := make(map[string]string)

	// Check Database
	sqlDB, err := uc.dbGorm.DB()
	if err != nil {
		status["db"] = "error: " + err.Error()
	} else {
		if err := sqlDB.Ping(); err != nil {
			status["db"] = "error: " + err.Error()
		} else {
			status["db"] = "connected"
		}
	}

	// Check Redis
	_, err = uc.Cache.Get(ctx, "health_check")
	if err != nil {
		// If key doesn't exist, that's fine - Redis is still working
		if err.Error() == "redis: nil" {
			status["redis"] = "connected"
		} else {
			status["redis"] = "error: " + err.Error()
		}
	} else {
		status["redis"] = "connected"
	}

	// Check MinIO
	if err := uc.Storage.HealthCheck(ctx); err != nil {
		status["minio"] = "error: " + err.Error()
	} else {
		status["minio"] = "connected"
	}

	return status, nil
}
