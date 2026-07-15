//go:generate apigen

package controller

import (
	"context"
	"net/http"

	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/usecase/therapist"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	dto "github.com/rdhmuhammad/phisiobook/shared/payload"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EmployeeController struct {
	base.BaseController
	uc EmployeeUsecase
}

type EmployeeUsecase interface {
	GetOnboardingList(ctx context.Context, query therapist.OnboardingListQuery) (dto.PaginationResponse[therapist.OnboardingListItem], error)
	GetOnboardingDetail(ctx context.Context, code string) (therapist.OnboardingDetailResponse, error)
	DeleteOnboarding(ctx context.Context, code string) error
}

func NewEmployeeController(dbConn *gorm.DB, port base.Port, controller base.BaseController) EmployeeController {
	return EmployeeController{
		BaseController: controller,
		uc:             therapist.NewUsecase(dbConn, port),
	}
}

func (ctrl EmployeeController) GetOnboardingList(c *gin.Context) {
	var request therapist.OnboardingListQuery
	if errs := ctrl.Enigma.BindQueryToFilterAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}
	request.SetIfEmpty()
	result, err := ctrl.uc.GetOnboardingList(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.GetOnboardingList.String()), err)
}

func (ctrl EmployeeController) GetOnboardingDetail(c *gin.Context) {
	code := c.Param("code")
	result, err := ctrl.uc.GetOnboardingDetail(c.Request.Context(), code)
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.GetOnboardingDetail.String()), err)
}

func (ctrl EmployeeController) DeleteOnboarding(c *gin.Context) {
	code := c.Param("code")
	err := ctrl.uc.DeleteOnboarding(c.Request.Context(), code)
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponseNoData(constant.DeleteOnboarding.String()), err)
}

func (ctrl EmployeeController) Route(router *gin.RouterGroup) {
	employeeGroup := router.Group("/employee")
	employeeGroup.Use(
		ctrl.Security.Validate(),
		ctrl.Security.Authorize(constant.RoleIsAdmin),
	)
	employeeGroup.GET("/onboarding", ctrl.GetOnboardingList)
	employeeGroup.GET("/onboarding/:code", ctrl.GetOnboardingDetail)
	employeeGroup.DELETE("/onboarding/:code", ctrl.DeleteOnboarding)
}