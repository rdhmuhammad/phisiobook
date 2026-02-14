package controller

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/usecase/booking"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/dto"
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type BookingController struct {
	BaseController
	usecase BookingUsecase
}

type BookingUsecase interface {
	CreateBooking(ctx context.Context, request booking.CreateBookingRequest) error
}

func NewBookingController(dbCache cache.Cache, dbConn *gorm.DB) BookingController {
	return BookingController{
		BaseController: NewBaseController(dbCache, dbConn),
	}
}

func (ctrl BookingController) CreateBooking(c *gin.Context) {
	var request booking.CreateBookingRequest
	if errs := ctrl.enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	err := ctrl.usecase.CreateBooking(c.Request.Context(), request)
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponseNoData(""), err)
}

func (r BookingController) Route(routeGr *gin.RouterGroup) {
	booking := routeGr.Group("/booking")

	booking.POST("/create",
		r.auth.Validate(),
		r.auth.Authorize(constant.RoleIsUser),
		r.idem.Idempotent(
			"/booking/create",
			"username",
			time.Millisecond*2,
		),
	)

}
