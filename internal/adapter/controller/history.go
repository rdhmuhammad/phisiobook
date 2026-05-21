//go:generate go run ../../../apitest_module/shared/runner.go

package controller

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/usecase/history"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	dto "github.com/rdhmuhammad/phisiobook/shared/payload"
	"gorm.io/gorm"
)

type HistoryController struct {
	base.BaseController
	uc HistoryUsecase
}

type HistoryUsecase interface {
	GetBookingHistory(ctx context.Context, request history.GetBookingHistoryRequest) (history.BookingHistoryResponse, error)
}

func NewHistoryController(dbConn *gorm.DB, controller base.BaseController, port base.Port) HistoryController {
	return HistoryController{
		BaseController: controller,
		uc:             history.NewUsecase(dbConn, port),
	}
}

func (ctrl HistoryController) GetBookingHistory(c *gin.Context) {
	var request history.GetBookingHistoryRequest
	if errs := ctrl.Enigma.BindQueryToFilterAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	result, err := ctrl.uc.GetBookingHistory(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(result, "Success"), err)
}

func (router HistoryController) Route(routes *gin.RouterGroup) {
	historyRoutes := routes.Group(
		"/history",
		router.Security.Validate(),
		router.Security.Authorize(constant.RoleIsUser),
	)

	historyRoutes.GET("/list", router.GetBookingHistory)
}
