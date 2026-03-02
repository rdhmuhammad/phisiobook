package controller

import (
	"context"

	"github.com/rdhmuhammad/phisiobook/internal/adapter/payload"
	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/usecase/booking"
	"github.com/rdhmuhammad/phisiobook/pkg/cache"
	"github.com/rdhmuhammad/phisiobook/pkg/cio"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	dto "github.com/rdhmuhammad/phisiobook/shared/payload"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zishang520/socket.io/servers/socket/v3"
	"gorm.io/gorm"
)

type BookingController struct {
	base.BaseController
	socketIO *cio.NS
	usecase  BookingUsecase
}

type BookingUsecase interface {
	UpdateStatus(ctx context.Context, request booking.UpdateStatus) (roomId string, err error)
	CreateBooking(ctx context.Context, request booking.CreateBookingRequest) error
}

func NewBookingController(dbCache cache.DbClient, dbConn *gorm.DB) BookingController {
	return BookingController{
		BaseController: base.NewBaseController(dbConn, dbCache),
	}
}

func (ctrl BookingController) CloseBooking(c *gin.Context) {
	var request booking.UpdateStatus
	if errs := ctrl.Enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}
	request.Code = c.Param("code")

	ctx := c.Request.Context()
	roomId, err := ctrl.usecase.UpdateStatus(ctx, request)
	if err != nil {
		ctrl.Mapper.ErrorResponse(c, err)
		return
	}

	userData := ctrl.Security.GetUserContext(ctx)

	err = ctrl.socketIO.Space.To(socket.Room(roomId)).
		Emit(payload.AlertError.String(),
			ctrl.Locale.GetLocalized(userData.Lang, constant.RoomNotValid))
	if err != nil {
		ctrl.Mapper.ErrorResponse(c, err)
		return
	}
	ctrl.socketIO.Space.To(socket.Room(roomId)).
		DisconnectSockets(true)

	ctrl.Mapper.NewResponse(c,
		dto.NewSuccessResponseNoData(constant.UpdateStatusBooking), err)
}

func (ctrl BookingController) CreateBooking(c *gin.Context) {
	var request booking.CreateBookingRequest
	if errs := ctrl.Enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	err := ctrl.usecase.CreateBooking(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponseNoData(""), err)
}

func (r BookingController) Route(routeGr *gin.RouterGroup) {
	booking := routeGr.Group("/booking")

	booking.POST("/create",
		r.Security.Validate(),
		r.Security.Authorize(constant.RoleIsUser),
		r.Idem.Idempotent(
			"/booking/create",
			"username",
			time.Millisecond*2,
		),
	)

}
