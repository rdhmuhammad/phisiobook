package controller

import (
	"base-be-golang/internal/adapter/payload"
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/usecase/booking"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/cio"
	"base-be-golang/pkg/dto"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zishang520/socket.io/servers/socket/v3"
	"gorm.io/gorm"
)

type BookingController struct {
	BaseController
	socketIO *cio.NS
	usecase  BookingUsecase
}

type BookingUsecase interface {
	UpdateStatus(ctx context.Context, request booking.UpdateStatus) (roomId string, err error)
	CreateBooking(ctx context.Context, request booking.CreateBookingRequest) error
}

func NewBookingController(dbCache cache.Cache, dbConn *gorm.DB) BookingController {
	return BookingController{
		BaseController: NewBaseController(dbCache, dbConn),
	}
}

func (ctrl BookingController) CloseBooking(c *gin.Context) {
	var request booking.UpdateStatus
	if errs := ctrl.enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}
	request.Code = c.Param("code")

	ctx := c.Request.Context()
	roomId, err := ctrl.usecase.UpdateStatus(ctx, request)
	if err != nil {
		ctrl.mapper.ErrorResponse(c, err)
		return
	}

	userData := ctrl.auth.GetSessionFromContext(ctx)

	err = ctrl.socketIO.Space.To(socket.Room(roomId)).
		Emit(payload.AlertError.String(),
			ctrl.localizer.GetLocalized(userData.Lang, constant.RoomNotValid))
	if err != nil {
		ctrl.mapper.ErrorResponse(c, err)
		return
	}
	ctrl.socketIO.Space.To(socket.Room(roomId)).
		DisconnectSockets(true)

	ctrl.mapper.NewResponse(c,
		dto.NewSuccessResponseNoData(constant.UpdateStatusBooking), err)
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
