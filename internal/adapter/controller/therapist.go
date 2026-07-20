//go:generate apigen

package controller

import (
	"context"
	"net/http"
	"path/filepath"
	"time"

	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/usecase/therapist"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	dto "github.com/rdhmuhammad/phisiobook/shared/payload"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TherapistController struct {
	base.BaseController
	uc TherapistUsecase
}

type TherapistUsecase interface {
	Register(ctx context.Context, request therapist.RegisterTherapistRequest) (therapist.RegisterTherapistResponse, error)
	Onboarding(ctx context.Context, request therapist.OnboardingRequest) (therapist.OnboardingResponse, error)
}

func NewTherapistController(dbConn *gorm.DB, port base.Port, controller base.BaseController) TherapistController {
	return TherapistController{
		BaseController: controller,
		uc:             therapist.NewUsecase(dbConn, port),
	}
}

func (ctrl TherapistController) Register(c *gin.Context) {
	var request therapist.RegisterTherapistRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorResponseWithMessage("Invalid input", err))
		return
	}

	profileHeader, err := c.FormFile("profile")
	if err == nil {
		profileFile, ferr := profileHeader.Open()
		if ferr != nil {
			ctrl.Mapper.ErrorResponse(c, ferr)
			return
		}
		defer profileFile.Close()
		request.Profile = therapist.FileInfo{
			Reader:    profileFile,
			Size:      profileHeader.Size,
			Extension: filepath.Ext(profileHeader.Filename),
		}
	}

	result, err := ctrl.uc.Register(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.RegisterTherapist.String()), err)
}

func (ctrl TherapistController) Onboarding(c *gin.Context) {
	ktpHeader, err := c.FormFile("ktpFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorResponseWithMessage("ktpFile is required", err))
		return
	}
	ktpFile, err := ktpHeader.Open()
	if err != nil {
		ctrl.Mapper.ErrorResponse(c, err)
		return
	}
	defer ktpFile.Close()

	sipHeader, err := c.FormFile("sipFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorResponseWithMessage("sipFile is required", err))
		return
	}
	sipFile, err := sipHeader.Open()
	if err != nil {
		ctrl.Mapper.ErrorResponse(c, err)
		return
	}
	defer sipFile.Close()

	strHeader, err := c.FormFile("strFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorResponseWithMessage("strFile is required", err))
		return
	}
	strFile, err := strHeader.Open()
	if err != nil {
		ctrl.Mapper.ErrorResponse(c, err)
		return
	}
	defer strFile.Close()

	ijazahHeader, err := c.FormFile("ijazahFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorResponseWithMessage("ijazahFile is required", err))
		return
	}
	ijazahFile, err := ijazahHeader.Open()
	if err != nil {
		ctrl.Mapper.ErrorResponse(c, err)
		return
	}
	defer ijazahFile.Close()

	request := therapist.OnboardingRequest{
		KtpFile:    therapist.FileInfo{Reader: ktpFile, Size: ktpHeader.Size, Extension: filepath.Ext(ktpHeader.Filename)},
		SipFile:    therapist.FileInfo{Reader: sipFile, Size: sipHeader.Size, Extension: filepath.Ext(sipHeader.Filename)},
		StrFile:    therapist.FileInfo{Reader: strFile, Size: strHeader.Size, Extension: filepath.Ext(strHeader.Filename)},
		IjazahFile: therapist.FileInfo{Reader: ijazahFile, Size: ijazahHeader.Size, Extension: filepath.Ext(ijazahHeader.Filename)},
		BankCode:   c.PostForm("bankCode"),
		AccName:    c.PostForm("accName"),
		AccNumber:  c.PostForm("accNumber"),
	}

	result, err := ctrl.uc.Onboarding(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.OnboardingSuccess.String()), err)
}

func (ctrl TherapistController) Route(router *gin.RouterGroup) {
	therapistGroup := router.Group("/therapist")

	therapistGroup.POST("/register",
		ctrl.Idem.Idempotent(
			"/therapist/register",
			"username",
			time.Millisecond*2,
		),
		ctrl.Register,
	)

	therapistGroup.POST("/onboarding",
		ctrl.Security.Validate(),
		ctrl.Security.Authorize(constant.RolesIsTerapis),
		ctrl.Onboarding,
	)
}
