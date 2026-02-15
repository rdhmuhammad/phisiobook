package controller

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/domain"
	"base-be-golang/internal/core/usecase/user_management"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/dto"
	"base-be-golang/pkg/logger"
	"base-be-golang/pkg/miniostorage"
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type UserManagementController struct {
	BaseController
	uc UserManagementUsecase
}

type UserManagementUsecase interface {
	DeleteUser(ctx context.Context, id uint) error
	UpsertUser(ctx context.Context, request user_management.CreateUserRequest, action int) error
	GetDetail(ctx context.Context, id uint) (user_management.UserDetailItem, error)
	GetList(ctx context.Context, query dto.GetListQueryNoPeriod) (dto.PaginationResponse[domain.UserListItem], error)
}

func NewUserManagementController(dbConn *gorm.DB, minio miniostorage.StorageMinio, dbCache cache.Cache, rz *logger.ReZero) UserManagementController {
	return UserManagementController{
		BaseController: NewBaseController(dbCache, dbConn),
		uc:             user_management.NewUsecase(dbConn, dbCache, minio, rz),
	}
}

func (ctrl UserManagementController) CreateUser(c *gin.Context) {
	var request user_management.CreateUserRequest
	if errs := ctrl.enigma.BindAndValidate(c, &request); errs != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	err := ctrl.uc.UpsertUser(c.Request.Context(), request, user_management.ActionIsCreateUser)
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponseNoData(constant.CreateUser), err)
}

func (ctrl UserManagementController) UpdateUser(c *gin.Context) {
	var request user_management.CreateUserRequest
	if errs := ctrl.enigma.BindAndValidate(c, &request); errs != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	userId, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorInvalidDataWithMessage(err.Error()))
		return
	}

	request.ID = uint(userId)
	err = ctrl.uc.UpsertUser(c.Request.Context(), request, user_management.ActionIsCreateUser)
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponseNoData(constant.UpdateUser), err)
}

func (ctrl UserManagementController) DeleteUser(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorInvalidDataWithMessage(err.Error()))
		return
	}

	err = ctrl.uc.DeleteUser(c.Request.Context(), uint(userId))
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponseNoData(constant.DeleteUser), err)
}

func (ctrl UserManagementController) GetDetailUser(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorInvalidDataWithMessage(err.Error()))
		return
	}

	result, err := ctrl.uc.GetDetail(c.Request.Context(), uint(userId))
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.GetDetailUser), err)
}

func (ctrl UserManagementController) GetListUser(c *gin.Context) {
	var request = dto.GetListQueryNoPeriod{}
	if errs := ctrl.enigma.BindQueryToFilterAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	request.SetIfEmpty()
	result, err := ctrl.uc.GetList(c.Request.Context(), request)
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.GetListUser), err)
}
