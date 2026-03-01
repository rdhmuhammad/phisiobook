package controller

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/port"
	"base-be-golang/internal/core/usecase/auth"
	"base-be-golang/pkg/dto"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	BaseController
	uc AuthUsecaseInterface
}

func NewAuthController(db *gorm.DB, controller BaseController, prt port.Port) AuthController {
	return AuthController{
		BaseController: controller,
		uc:             auth.NewUsecase(db, prt),
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

func (ctrl AuthController) Register(c *gin.Context, roleName string) {
	var request auth.RegisterRequest
	if errs := ctrl.enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	request.RoleName = roleName
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
	routeAuth := router.Group("/auth-user")

	// ==================== USER ROUTE ====================
	userRoute := routeAuth.Group("/user")
	userRoute.POST("/register",
		ctrl.idem.Idempotent(
			"/register/user",
			"username",
			time.Millisecond*2,
		),
		func(c *gin.Context) {
			ctrl.Register(c, constant.RoleIsUser)
		},
	)
	userRoute.POST("/logout",
		ctrl.auth.Validate(),
		ctrl.auth.Authorize(constant.RoleIsUser),
		func(c *gin.Context) {
			ctrl.Logout(c, constant.ContextMobile)
		})
	userRoute.POST("/login",
		func(c *gin.Context) {
			ctrl.Login(c, constant.ContextMobile)
		},
	)
	// =========================================================

	// ==================== THERAPIST ROUTE ====================
	therapyRoute := router.Group("/therapist")
	therapyRoute.POST("/register",
		ctrl.idem.Idempotent(
			"/register/therapist",
			"username",
			time.Millisecond*2,
		),
		func(c *gin.Context) {
			ctrl.Register(c, constant.RolesIsTerapis)
		},
	)
	therapyRoute.POST("/logout",
		ctrl.auth.Validate(),
		ctrl.auth.Authorize(constant.RolesIsTerapis),
		func(c *gin.Context) {
			ctrl.Logout(c, constant.ContextDashboard)
		})
	therapyRoute.POST("/login",
		func(c *gin.Context) {
			ctrl.Login(c, constant.ContextDashboard)
		},
	)

	// ==================== DASHBOARD ROUTE ====================
	dashboardAuth := router.Group("/employee")
	dashboardAuth.POST("/register",
		ctrl.idem.Idempotent(
			"/register/terapis",
			"username",
			time.Millisecond*2,
		),
		func(c *gin.Context) {
			ctrl.Register(c, constant.RolesIsTerapis)
		},
	)

	dashboardAuth.POST("/login",
		func(c *gin.Context) {
			ctrl.Login(c, constant.ContextDashboard)
		},
	)
	dashboardAuth.POST("/logout",
		ctrl.auth.Validate(),
		ctrl.auth.Authorize(constant.RolesIsTerapis, constant.RoleIsAdmin),
		func(c *gin.Context) {
			ctrl.Logout(c, constant.ContextDashboard)
		})
	// ===========================================================

	routeAuth.POST(
		"/verify-acc",
		ctrl.VerifyAcc,
	)

	routeAuth.POST(
		"/resend-otp",
		ctrl.idem.Idempotent(
			"/register",
			"username",
			time.Minute*1,
		),
		ctrl.ResendOTP,
	)
}
