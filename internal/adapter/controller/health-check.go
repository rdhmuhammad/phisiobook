package controller

import (
	"context"
	"github.com/rdhmuhammad/phisiobook/internal/core/usecase/health"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	"github.com/rdhmuhammad/phisiobook/shared/payload"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthCheckController struct {
	base.BaseController
	uc HealthCheckUsecase
}

type HealthCheckUsecase interface {
	CheckHealth(ctx context.Context) (map[string]string, error)
}

func NewHealthController(dbConn *gorm.DB, controller base.BaseController, port base.Port) HealthCheckController {
	return HealthCheckController{
		BaseController: controller,
		uc:             health.New(dbConn, port),
	}
}

func (ctrl HealthCheckController) HealthCheck(c *gin.Context) {

	res, err := ctrl.uc.CheckHealth(c.Request.Context())
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponse(res, "Success"), err)
}

func (router HealthCheckController) Route(routes *gin.RouterGroup) {
	healthRoutes := routes.Group("/health")
	healthRoutes.GET("/status", router.HealthCheck)
}
