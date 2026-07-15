//go:generate apigen

package controller

import (
	"context"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/usecase/homepage"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	dto "github.com/rdhmuhammad/phisiobook/shared/payload"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HomepageController struct {
	base.BaseController
	port base.Port
	uc   HomepageUsecase
}

type HomepageUsecase interface {
	GetSummaryHome(ctx context.Context) (*homepage.SummaryHomeResponse, error)
	GetCityDropdown(ctx context.Context) ([]homepage.CityDropdownResponse, error)
	GetTherapist(ctx context.Context, cityId uint) ([]homepage.TherapistDropdownResponse, error)
}

func NewHomepageController(dbConn *gorm.DB, controller base.BaseController, port base.Port) HomepageController {
	return HomepageController{
		BaseController: controller,
		port:           port,
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

func (ctrl HomepageController) GetTherapistDropdown(c *gin.Context) {
	cityId, err := strconv.ParseUint(c.Param("cityId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.DefaultErrorResponse(err))
		return
	}
	res, err := ctrl.uc.GetTherapist(c.Request.Context(), uint(cityId))
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(res, constant.DropdownTherapistSuccess.String()), err)
}

func (ctrl HomepageController) DownloadFile(c *gin.Context) {
	fileName := c.Query("fileName")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorResponseWithMessage("fileName is required", nil))
		return
	}
	buf, err := ctrl.port.Storage.GetFile(c.Request.Context(), fileName)
	if err != nil {
		ctrl.Mapper.ErrorResponse(c, err)
		return
	}
	ext := strings.ToLower(filepath.Ext(fileName))
	contentType := "application/octet-stream"
	switch ext {
	case ".png":
		contentType = constant.MIMEPNG
	case ".jpg", ".jpeg":
		contentType = constant.MIMEJPEG
	case ".pdf":
		contentType = constant.MIMEPDF
	case ".xlsx":
		contentType = constant.MIMEXLSX
	}
	c.Data(http.StatusOK, contentType, buf.Bytes())
}

func (router HomepageController) Route(routes *gin.RouterGroup) {
	homepageRoutes := routes.Group("/homepage")
	homepageRoutes.GET("/summary", router.GetSummaryHome)
	homepageRoutes.GET("/cities-dropdown", router.GetCityDropdown)
	homepageRoutes.GET("/therapist-dropdown/:cityId", router.GetTherapistDropdown)
	routes.GET("/download", router.DownloadFile)
}