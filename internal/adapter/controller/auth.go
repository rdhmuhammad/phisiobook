package controller

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/usecase/auth"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/dto"
	"base-be-golang/pkg/miniostorage"
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type AuthController struct {
	BaseController
	uc AuthUsecaseInterface
}

func NewAuthController(db *gorm.DB, cacheDb cache.Cache, conn miniostorage.StorageMinio) AuthController {
	return AuthController{
		BaseController: NewBaseController(cacheDb, db),
		uc:             auth.NewUsecase(db, cacheDb, conn),
	}
}

type AuthUsecaseInterface interface {
	Register(ctx context.Context, request auth.RegisterRequest) (auth.RegisterResponse, error)
	Logout(ctx context.Context, role string) error
	Login(ctx context.Context, request auth.LoginRequest) (auth.LoginResponse, error)
	VerifyAcc(ctx context.Context, request auth.VerifyAccRequest) (auth.VerifyAccResponse, error)
	ResendOTP(ctx context.Context, request auth.SendOtpRequest) error
}

func (ctrl AuthController) Logout(c *gin.Context, role string) {
	err := ctrl.uc.Logout(c.Request.Context(), role)
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponseNoData(constant.LogoutSuccess), err)
}

func (ctrl AuthController) Login(c *gin.Context, role string) {
	var request auth.LoginRequest
	if errs := ctrl.enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}
	request.Role = role
	result, err := ctrl.uc.Login(c.Request.Context(), request)
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.LoginSuccess), err)
}

func (ctrl AuthController) Register(c *gin.Context) {
	var request auth.RegisterRequest
	if errs := ctrl.enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	result, err := ctrl.uc.Register(c.Request.Context(), request)
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.RegisterSuccess), err)
}

func (ctrl AuthController) VerifyAcc(c *gin.Context) {
	var request auth.VerifyAccRequest
	if errs := ctrl.enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	result, err := ctrl.uc.VerifyAcc(c.Request.Context(), request)
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.VerifyOtpSuccess), err)
}

func (ctrl AuthController) ResendOTP(c *gin.Context) {
	var request auth.SendOtpRequest
	if errs := ctrl.enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
	}

	err := ctrl.uc.ResendOTP(c.Request.Context(), request)
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponseNoData(constant.ResendOtpSuccess), err)
}

func (ctrl AuthController) Route(router *gin.RouterGroup) {
	userAuth := router.Group("/auth-user")
	userAuth.POST("/register",
		ctrl.idem.Idempotent(
			"/register",
			"username",
			time.Millisecond*2,
		),
		ctrl.Register,
	)

	userAuth.POST("/login",
		func(c *gin.Context) {
			ctrl.Login(c, constant.ContextMobile)
		},
	)

	userAuth.POST("/login/admin",
		func(c *gin.Context) {
			ctrl.Login(c, constant.ContextDashboard)
		},
	)

	userAuth.POST("/logout",
		ctrl.auth.Validate(),
		ctrl.auth.Authorize(constant.RoleIsAdmin, constant.RoleIsUser),
		func(c *gin.Context) {
			ctrl.Logout(c, constant.ContextMobile)
		})

	userAuth.POST("/logout/admin",
		ctrl.auth.Validate(),
		ctrl.auth.Authorize(constant.RolesIsMobile),
		func(c *gin.Context) {
			ctrl.Logout(c, constant.ContextDashboard)
		})

	userAuth.POST(
		"/verify-acc",
		ctrl.VerifyAcc,
	)

	userAuth.POST(
		"/resend-otp",
		ctrl.idem.Idempotent(
			"/register",
			"username",
			time.Minute*1,
		),
		ctrl.ResendOTP,
	)
}
