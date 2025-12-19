package controller

import (
	"base-be-golang/internal/core/usecase/health"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/dto"
	"base-be-golang/pkg/miniostorage"
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthCheckController struct {
	BaseController
	uc HealthCheckUsecase
}

type HealthCheckUsecase interface {
	CheckHealth(ctx context.Context) (map[string]string, error)
}

func NewHealthController(dbCache cache.Cache, dbConn *gorm.DB, minioConn miniostorage.StorageMinio) HealthCheckController {
	return HealthCheckController{
		BaseController: NewBaseController(dbCache, dbConn),
		uc:             health.New(dbConn, dbCache, minioConn),
	}
}

func (ctrl HealthCheckController) HealthCheck(c *gin.Context) {

	res, err := ctrl.uc.CheckHealth(c.Request.Context())
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(res, "Success"), err)
}

func (router HealthCheckController) Route(routes *gin.RouterGroup) {
	healthRoutes := routes.Group("/health")
	healthRoutes.GET("/status", router.HealthCheck)
}
