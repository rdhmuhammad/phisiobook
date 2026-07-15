//go:generate apigen

package controller

import (
	"context"
	"net/http"
	"strconv"

	"github.com/rdhmuhammad/phisiobook/shared/base"
	"github.com/rdhmuhammad/phisiobook/shared/payload"

	"iam_module/internal/adapter/repository"
	"iam_module/internal/core/domain"
	user_management "iam_module/internal/core/usecase/usermanagement"
	"iam_module/shared/constant"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserManagementController struct {
	base.BaseController
	uc UserManagementUsecase
}

type UserManagementUsecase interface {
	DeleteUser(ctx context.Context, id uint) error
	UpsertUser(ctx context.Context, request user_management.CreateUserRequest, action int) error
	GetDetail(ctx context.Context, id uint) (user_management.UserDetailItem, error)
	GetList(ctx context.Context, query repository.UserListQuery) (payload.PaginationResponse[domain.UserListItem], error)
}

func NewUserManagementController(dbConn *gorm.DB, port base.Port, controller base.BaseController) UserManagementController {
	return UserManagementController{
		BaseController: controller,
		uc:             user_management.NewUsecase(dbConn, port),
	}
}

func (ctrl UserManagementController) CreateUser(c *gin.Context) {
	var request user_management.CreateUserRequest
	if errs := ctrl.Enigma.BindAndValidate(c, &request); errs != nil {
		c.JSON(http.StatusBadRequest, payload.DefaultInvalidInputFormResponse(errs))
		return
	}

	err := ctrl.uc.UpsertUser(c.Request.Context(), request, user_management.ActionIsCreateUser)
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponseNoData(constant.CreateUser.String()), err)
}

func (ctrl UserManagementController) UpdateUser(c *gin.Context) {
	var request user_management.CreateUserRequest
	if errs := ctrl.Enigma.BindAndValidate(c, &request); errs != nil {
		c.JSON(http.StatusBadRequest, payload.DefaultInvalidInputFormResponse(errs))
		return
	}

	userId, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, payload.DefaultErrorInvalidDataWithMessage(err.Error()))
		return
	}

	request.ID = uint(userId)
	err = ctrl.uc.UpsertUser(c.Request.Context(), request, user_management.ActionIsCreateUser)
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponseNoData(constant.UpdateUser.String()), err)
}

func (ctrl UserManagementController) DeleteUser(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, payload.DefaultErrorInvalidDataWithMessage(err.Error()))
		return
	}

	err = ctrl.uc.DeleteUser(c.Request.Context(), uint(userId))
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponseNoData(constant.DeleteUser.String()), err)
}

func (ctrl UserManagementController) GetDetailUser(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, payload.DefaultErrorInvalidDataWithMessage(err.Error()))
		return
	}

	result, err := ctrl.uc.GetDetail(c.Request.Context(), uint(userId))
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponse(result, constant.GetDetailUser.String()), err)
}

func (ctrl UserManagementController) GetListUser(c *gin.Context) {
	var request = repository.UserListQuery{
		Filter: &payload.GetListQueryNoPeriod{},
	}
	if errs := ctrl.Enigma.BindQueryToFilterAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, payload.DefaultInvalidInputFormResponse(errs))
		return
	}

	request.Filter.SetIfEmpty()
	result, err := ctrl.uc.GetList(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponse(result, constant.GetListUser.String()), err)
}

func (ctrl UserManagementController) Route(handler *gin.RouterGroup) {

}
