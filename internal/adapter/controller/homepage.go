package controller

import (
	"base-be-golang/internal/core/usecase/homepage"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/dto"
	"base-be-golang/pkg/miniostorage"
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HomepageController struct {
	BaseController
	uc HomepageUsecase
}

type HomepageUsecase interface {
	GetSummaryHome(ctx context.Context) (*homepage.SummaryHomeResponse, error)
}

func NewHomepageController(dbCache cache.Cache, dbConn *gorm.DB, minioConn miniostorage.StorageMinio) HomepageController {
	return HomepageController{
		BaseController: NewBaseController(dbCache, dbConn),
		uc:             homepage.New(dbConn, dbCache, minioConn),
	}
}

func (ctrl HomepageController) GetSummaryHome(c *gin.Context) {
	res, err := ctrl.uc.GetSummaryHome(c.Request.Context())
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(res, "Success"), err)
}

func (router HomepageController) Route(routes *gin.RouterGroup) {
	homepageRoutes := routes.Group("/homepage")
	homepageRoutes.GET("/summary", router.GetSummaryHome)
}
