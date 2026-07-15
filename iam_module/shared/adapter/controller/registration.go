//go:generate apigen

package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/rdhmuhammad/phisiobook/shared/base"
	"github.com/rdhmuhammad/phisiobook/shared/payload"

	"iam_module/internal/core/constant"
	"iam_module/internal/core/usecase/registration"
	constant2 "iam_module/shared/constant"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	base.BaseController
	uc AuthUsecaseInterface
}

func NewAuthController(dbConn *gorm.DB, port base.Port, controller base.BaseController) AuthController {
	return AuthController{
		BaseController: controller,
		uc:             registration.NewUsecase(dbConn, port),
	}
}

type AuthUsecaseInterface interface {
	Register(ctx context.Context, request registration.RegisterRequest) (registration.RegisterResponse, error)
	Logout(ctx context.Context, role string) error
	Login(ctx context.Context, request registration.LoginRequest) (registration.LoginResponse, error)
	VerifyAcc(ctx context.Context, request registration.VerifyAccRequest) (registration.VerifyAccResponse, error)
	ResendOTP(ctx context.Context, request registration.SendOtpRequest) error
}

func (ctrl AuthController) Logout(c *gin.Context, role string) {
	err := ctrl.uc.Logout(c.Request.Context(), role)
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponseNoData(constant2.LogoutSuccess.String()), err)
}

func (ctrl AuthController) Login(c *gin.Context, role string) {
	var request registration.LoginRequest
	if errs := ctrl.Enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, payload.DefaultInvalidInputFormResponse(errs))
		return
	}
	request.Role = role
	result, err := ctrl.uc.Login(c.Request.Context(), request)
	c.Set(string(constant2.FallBackLangLogin), result.Lang)
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponse(result, constant2.LoginSuccess.String()), err)
}

func (ctrl AuthController) LoginUser(c *gin.Context) {
	ctrl.Login(c, constant.ContextMobile)
}

func (ctrl AuthController) LoginAdmin(c *gin.Context) {
	ctrl.Login(c, constant.ContextDashboard)
}

func (ctrl AuthController) LoginTherapist(c *gin.Context) {
	ctrl.Login(c, constant.ContextTherapist)
}

func (ctrl AuthController) LogoutUser(c *gin.Context) {
	ctrl.Logout(c, constant.ContextMobile)
}

func (ctrl AuthController) LogoutAdmin(c *gin.Context) {
	ctrl.Logout(c, constant.ContextDashboard)
}

func (ctrl AuthController) Register(c *gin.Context) {
	var request registration.RegisterRequest
	if errs := ctrl.Enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, payload.DefaultInvalidInputFormResponse(errs))
		return
	}

	result, err := ctrl.uc.Register(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponse(result, constant2.RegisterSuccess.String()), err)
}

func (ctrl AuthController) VerifyAcc(c *gin.Context) {
	var request registration.VerifyAccRequest
	if errs := ctrl.Enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, payload.DefaultInvalidInputFormResponse(errs))
		return
	}

	result, err := ctrl.uc.VerifyAcc(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponse(result, constant2.VerifyOtpSuccess.String()), err)
}

func (ctrl AuthController) ResendOTP(c *gin.Context) {
	var request registration.SendOtpRequest
	if errs := ctrl.Enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, payload.DefaultInvalidInputFormResponse(errs))
	}

	err := ctrl.uc.ResendOTP(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponseNoData(constant2.ResendOtpSuccess.String()), err)
}

func (ctrl AuthController) Route(router *gin.RouterGroup) {
	userAuth := router.Group("/auth-user")

	userAuth.POST("/register",
		ctrl.Idem.Idempotent(
			"/register",
			"username",
			time.Millisecond*2,
		),
		ctrl.Register,
	)

	userAuth.POST("/user/login", ctrl.LoginUser)

	userAuth.POST("/login/admin", ctrl.LoginAdmin)

	userAuth.POST("/login/therapist", ctrl.LoginTherapist)

	userAuth.POST("/logout",
		ctrl.Security.Validate(),
		ctrl.Security.Authorize(constant.RoleIsAdmin, constant.RoleIsUser),
		ctrl.LogoutUser,
	)

	userAuth.POST("/logout/admin",
		ctrl.Security.Validate(),
		ctrl.Security.Authorize(constant.RolesIsMobile),
		ctrl.LogoutAdmin,
	)

	userAuth.POST("/verify-acc", ctrl.VerifyAcc)

	userAuth.POST("/resend-otp",
		ctrl.Idem.Idempotent(
			"/resend-otp",
			"username",
			time.Minute*1,
		),
		ctrl.ResendOTP,
	)
}
