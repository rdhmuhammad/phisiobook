package controller

import (
	"base-be-golang/internal/core/port"
	"base-be-golang/internal/core/usecase/homepage"
	"base-be-golang/pkg/dto"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HomepageController struct {
	BaseController
	uc HomepageUsecase
}

type HomepageUsecase interface {
	GetSummaryHome(ctx context.Context) (*homepage.SummaryHomeResponse, error)
	GetCityDropdown(ctx context.Context) ([]homepage.CityDropdownResponse, error)
	GetTherapist(ctx context.Context, cityId uint) ([]homepage.TherapistDropdownResponse, error)
}

func NewHomepageController(dbConn *gorm.DB, controller BaseController, port port.Port) HomepageController {
	return HomepageController{
		BaseController: controller,
		uc:             homepage.New(dbConn, port),
	}
}

func (ctrl HomepageController) GetSummaryHome(c *gin.Context) {
	res, err := ctrl.uc.GetSummaryHome(c.Request.Context())
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(res, "Success"), err)
}

func (ctrl HomepageController) GetTherapistDropdown(c *gin.Context) {
	cityId, err := strconv.ParseUint(c.Param("cityId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.DefaultErrorResponse(err))
		return
	}

	res, err := ctrl.uc.GetTherapist(c.Request.Context(), uint(cityId))
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(res, "Success"), err)
}

func (ctrl HomepageController) GetCityDropdown(c *gin.Context) {
	res, err := ctrl.uc.GetCityDropdown(c.Request.Context())
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(res, "Success"), err)
}

func (router HomepageController) Route(routes *gin.RouterGroup) {
	homepageRoutes := routes.Group("/homepage")
	homepageRoutes.GET("/summary", router.GetSummaryHome)
	homepageRoutes.GET("/cities-dropdown", router.GetCityDropdown)
	homepageRoutes.GET("/therapist-dropdown/:cityId", router.GetTherapistDropdown)
}
