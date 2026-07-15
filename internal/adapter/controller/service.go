//go:generate apigen

package controller

import (
	"context"

	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/usecase/service"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	dto "github.com/rdhmuhammad/phisiobook/shared/payload"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ServiceController struct {
	base.BaseController
	uc ServiceUsecase
}

type ServiceUsecase interface {
	CreateService(ctx context.Context, request service.CreateServiceRequest) (service.ServiceDetailResponse, error)
	UpdateService(ctx context.Context, request service.UpdateServiceRequest) (service.ServiceDetailResponse, error)
	DeleteService(ctx context.Context, id uint) error
	GetServiceDetail(ctx context.Context, id uint) (service.ServiceDetailResponse, error)
	GetServiceList(ctx context.Context, query dto.GetListQueryNoPeriod) (dto.PaginationResponse[service.ServiceListItem], error)
	GetCategories(ctx context.Context) ([]service.CategoryItem, error)
}

func NewServiceController(dbConn *gorm.DB, prt base.Port, ctrl base.BaseController) ServiceController {
	return ServiceController{
		BaseController: ctrl,
		uc:             service.NewUsecase(dbConn, prt),
	}
}

func (ctrl ServiceController) CreateService(c *gin.Context) {
	var request service.CreateServiceRequest
	if errs := ctrl.Enigma.BindAndValidate(c, &request); errs != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	result, err := ctrl.uc.CreateService(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.CreateService.String()), err)
}

func (ctrl ServiceController) UpdateService(c *gin.Context) {
	var request service.UpdateServiceRequest
	if errs := ctrl.Enigma.BindAndValidate(c, &request); errs != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	serviceID, err := strconv.ParseUint(c.Param("serviceId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorInvalidDataWithMessage("Invalid service ID"))
		return
	}

	request.ID = uint(serviceID)
	result, err := ctrl.uc.UpdateService(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.UpdateService.String()), err)
}

func (ctrl ServiceController) DeleteService(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("serviceId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorInvalidDataWithMessage("Invalid service ID"))
		return
	}

	err = ctrl.uc.DeleteService(c.Request.Context(), uint(serviceID))
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponseNoData(constant.DeleteService.String()), err)
}

func (ctrl ServiceController) GetServiceDetail(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("serviceId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorInvalidDataWithMessage("Invalid service ID"))
		return
	}

	result, err := ctrl.uc.GetServiceDetail(c.Request.Context(), uint(serviceID))
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.GetDetailService.String()), err)
}

func (ctrl ServiceController) GetServiceList(c *gin.Context) {
	var request = dto.GetListQueryNoPeriod{}
	if errs := ctrl.Enigma.BindQueryToFilterAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	request.SetIfEmpty()
	result, err := ctrl.uc.GetServiceList(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.GetListService.String()), err)
}

func (ctrl ServiceController) GetCategories(c *gin.Context) {
	result, err := ctrl.uc.GetCategories(c.Request.Context())
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.GetCategories.String()), err)
}

func (ctrl ServiceController) Route(router *gin.RouterGroup) {
	serviceGroup := router.Group("/service")

	// All routes require admin authentication
	serviceGroup.Use(ctrl.Security.Validate(), ctrl.Security.Authorize(constant.RoleIsAdmin))

	// CRUD endpoints
	serviceGroup.POST("", ctrl.CreateService)
	serviceGroup.GET("", ctrl.GetServiceList)
	serviceGroup.GET("/categories", ctrl.GetCategories)
	serviceGroup.GET("/:serviceId", ctrl.GetServiceDetail)
	serviceGroup.PUT("/:serviceId", ctrl.UpdateService)
	serviceGroup.DELETE("/:serviceId", ctrl.DeleteService)

}
