//go:generate go run ../../../apitest_module/shared/runner.go

package controller

import (
	"context"
	"strconv"

	"github.com/rdhmuhammad/phisiobook/internal/adapter/payload"
	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/usecase/booking"
	"github.com/rdhmuhammad/phisiobook/pkg/cio"
	"github.com/rdhmuhammad/phisiobook/pkg/mongodb"
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
	GetAdjustPrice(ctx context.Context, therapistCode string) (booking.AdjustPriceResponse, error)
	RescheduleBooking(ctx context.Context, request booking.RescheduleBookingRequest) (booking.RescheduleBookingResponse, error)
	GetCityDropdown(ctx context.Context) ([]booking.CityDropdownResponse, error)
	GetTherapist(ctx context.Context, cityId uint) ([]booking.TherapistDropdownResponse, error)
}

func NewBookingController(dbConn *gorm.DB, mongoConn *mongodb.Conn, prt base.Port, ctrl base.BaseController) BookingController {
	return BookingController{
		BaseController: ctrl,
		usecase:        booking.NewUsecase(dbConn, mongoConn, prt),
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
		Emit(payload.Alert_error.Topic(),
			ctrl.Locale.GetLocalized(userData.Lang, constant.RoomNotValid.String()))
	if err != nil {
		ctrl.Mapper.ErrorResponse(c, err)
		return
	}
	ctrl.socketIO.Space.To(socket.Room(roomId)).
		DisconnectSockets(true)

	ctrl.Mapper.NewResponse(c,
		dto.NewSuccessResponseNoData(constant.UpdateStatusBooking.String()), err)
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

func (ctrl BookingController) GetTherapistDropdown(c *gin.Context) {
	cityId, err := strconv.ParseUint(c.Param("cityId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.DefaultErrorResponse(err))
		return
	}

	res, err := ctrl.usecase.GetTherapist(c.Request.Context(), uint(cityId))
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(res, constant.DropdownTherapistSuccess.String()), err)
}

func (ctrl BookingController) GetCityDropdown(c *gin.Context) {
	res, err := ctrl.usecase.GetCityDropdown(c.Request.Context())
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(res, constant.DropdownCitySuccess.String()), err)
}

func (ctrl BookingController) GetAdjustPrice(c *gin.Context) {
	therapistCode := c.Param("therapistCode")
	res, err := ctrl.usecase.GetAdjustPrice(c.Request.Context(), therapistCode)
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(res, constant.DropdownTherapistSuccess.String()), err)
}

func (ctrl BookingController) RescheduleBooking(c *gin.Context) {
	var request booking.RescheduleBookingRequest
	if errs := ctrl.Enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, dto.DefaultInvalidInputFormResponse(errs))
		return
	}

	request.Code = c.Param("code")
	res, err := ctrl.usecase.RescheduleBooking(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(res, constant.RescheduleBookingSuccess), err)
}

func (r BookingController) Route(routeGr *gin.RouterGroup) {
	bookingRouter := routeGr.Group("/booking")

	bookingRouter.POST("/create",
		r.Security.Validate(),
		r.Security.Authorize(constant.RoleIsUser),
		r.Idem.Idempotent(
			"/bookingRouter/create",
			"username",
			time.Millisecond*2,
		),
		r.CreateBooking,
	)
	bookingRouter.GET("/dropdown-cities", r.GetCityDropdown)
	bookingRouter.GET("/dropdown-therapist/:cityId", r.GetTherapistDropdown)
	bookingRouter.GET("/adjust-price/:therapistCode", r.GetAdjustPrice)
	bookingRouter.PUT("/reschedule/:code",
		r.Security.Validate(),
		r.Security.Authorize(constant.RoleIsUser),
		r.RescheduleBooking,
	)

}
