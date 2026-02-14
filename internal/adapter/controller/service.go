package controller

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/usecase/service"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/dto"
	"base-be-golang/pkg/miniostorage"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ServiceController struct {
	BaseController
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

func NewServiceController(dbConn *gorm.DB, minio miniostorage.StorageMinio, dbCache cache.Cache) ServiceController {
	return ServiceController{
		BaseController: NewBaseController(dbCache, dbConn),
		uc:             service.NewUsecase(dbConn, dbCache, minio),
	}
}

func (ctrl ServiceController) CreateService(c *gin.Context) {
	var request service.CreateServiceRequest
	if errs := ctrl.enigma.BindAndValidate(c, &request); errs != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	result, err := ctrl.uc.CreateService(c.Request.Context(), request)
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.CreateService), err)
}

func (ctrl ServiceController) UpdateService(c *gin.Context) {
	var request service.UpdateServiceRequest
	if errs := ctrl.enigma.BindAndValidate(c, &request); errs != nil {
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
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.UpdateService), err)
}

func (ctrl ServiceController) DeleteService(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("serviceId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorInvalidDataWithMessage("Invalid service ID"))
		return
	}

	err = ctrl.uc.DeleteService(c.Request.Context(), uint(serviceID))
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponseNoData(constant.DeleteService), err)
}

func (ctrl ServiceController) GetServiceDetail(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("serviceId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorInvalidDataWithMessage("Invalid service ID"))
		return
	}

	result, err := ctrl.uc.GetServiceDetail(c.Request.Context(), uint(serviceID))
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.GetDetailService), err)
}

func (ctrl ServiceController) GetServiceList(c *gin.Context) {
	var request = dto.GetListQueryNoPeriod{}
	if errs := ctrl.enigma.BindQueryToFilterAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	request.SetIfEmpty()
	result, err := ctrl.uc.GetServiceList(c.Request.Context(), request)
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.GetListService), err)
}

func (ctrl ServiceController) GetCategories(c *gin.Context) {
	result, err := ctrl.uc.GetCategories(c.Request.Context())
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.GetCategories), err)
}

func (ctrl ServiceController) Route(router *gin.RouterGroup) {
	serviceGroup := router.Group("/service")
	{
		// All routes require admin authentication
		serviceGroup.Use(ctrl.auth.Validate(), ctrl.auth.Authorize(constant.RoleIsAdmin))

		// CRUD endpoints
		serviceGroup.POST("", ctrl.CreateService)
		serviceGroup.GET("", ctrl.GetServiceList)
		serviceGroup.GET("/categories", ctrl.GetCategories)
		serviceGroup.GET("/:serviceId", ctrl.GetServiceDetail)
		serviceGroup.PUT("/:serviceId", ctrl.UpdateService)
		serviceGroup.DELETE("/:serviceId", ctrl.DeleteService)
	}
}
