//go:generate go run ../../../apitest_module/shared/runner.go

package controller

import (
	"context"

	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/usecase/homepage"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	dto "github.com/rdhmuhammad/phisiobook/shared/payload"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HomepageController struct {
	base.BaseController
	uc HomepageUsecase
}

type HomepageUsecase interface {
	GetSummaryHome(ctx context.Context) (*homepage.SummaryHomeResponse, error)
	GetCityDropdown(ctx context.Context) ([]homepage.CityDropdownResponse, error)
}

func NewHomepageController(dbConn *gorm.DB, controller base.BaseController, port base.Port) HomepageController {
	return HomepageController{
		BaseController: controller,
		uc:             homepage.New(dbConn, port),
	}
}

func (ctrl HomepageController) GetSummaryHome(c *gin.Context) {
	res, err := ctrl.uc.GetSummaryHome(c.Request.Context())
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(res, "Success"), err)
}

func (ctrl HomepageController) GetCityDropdown(c *gin.Context) {
	res, err := ctrl.uc.GetCityDropdown(c.Request.Context())
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(res, constant.DropdownCitySuccess.String()), err)
}

func (router HomepageController) Route(routes *gin.RouterGroup) {
	homepageRoutes := routes.Group("/homepage")
	homepageRoutes.GET("/summary", router.GetSummaryHome)
	homepageRoutes.GET("/cities-dropdown", router.GetCityDropdown)
}
