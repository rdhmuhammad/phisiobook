package socket

import (
	"base-be-golang/internal/adapter/controller"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/environment"
	"base-be-golang/pkg/mapper"
	"base-be-golang/pkg/middleware"

	"gorm.io/gorm"
)

type BaseSocket struct {
	auth   controller.AuthInterface
	enigma controller.EnigmaUtility
	mapper controller.MapperUtility
	idem   controller.Idempotent
	env    controller.Environment
}

func NewBaseSocket(cache cache.Cache, dbConn *gorm.DB) BaseSocket {
	return BaseSocket{
		auth:   middleware.NewAuth(dbConn),
		enigma: middleware.NewEnigma(),
		mapper: mapper.NewMapper(),
		env:    environment.NewEnvironment(),
		idem:   middleware.NewIdempotent(cache),
	}
}
