package controller

import (
	"base-be-golang/internal/core/port"
	"base-be-golang/internal/core/usecase/health"
	"base-be-golang/pkg/dto"
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

func NewHealthController(dbConn *gorm.DB, controller BaseController, port port.Port) HealthCheckController {
	return HealthCheckController{
		BaseController: controller,
		uc:             health.New(dbConn, port),
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
