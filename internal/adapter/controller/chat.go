package controller

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/port"
	"base-be-golang/internal/core/usecase/caching_chat"
	"base-be-golang/pkg/dto"
	"base-be-golang/pkg/mongodb"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChatController struct {
	BaseController
	cachedUc CachedChatUsecase
}

func NewChatController(mongoConn *mongodb.Conn, controller BaseController, port port.Port) ChatController {
	return ChatController{
		cachedUc:       caching_chat.NewUsecase(mongoConn, port),
		BaseController: controller,
	}
}

type CachedChatUsecase interface {
	CacheChat(ctx context.Context, request caching_chat.CacheChatRequest)
	CacheRoom(ctx context.Context, request caching_chat.CacheRoomRequest, cancle context.CancelFunc) (caching_chat.CacheRoomResponse, error)
	GetCached(ctx context.Context, request caching_chat.GetCachedRequest) (dto.PaginationResponse[caching_chat.CachedChat], error)
}

func (ctrl ChatController) GetCached(c *gin.Context) {
	var request = caching_chat.GetCachedRequest{
		Filter: &dto.GetListQueryNoPeriod{},
	}
	if errs := ctrl.enigma.BindQueryToFilter(c, &request); errs != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorInvalidDataWithMessage(errs.Error()))
		return
	}

	request.Filter.SetIfEmpty()
	result, err := ctrl.cachedUc.GetCached(c.Request.Context(), request)
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(result, ""), err)
}

func (ctrl ChatController) Route(route *gin.RouterGroup) {
	restRouter := route.Group(
		"",
		ctrl.auth.Validate(),
		ctrl.auth.Authorize(constant.RoleIsUser, constant.RolesIsTerapis))
	restRouter.GET("/history", ctrl.GetCached)
}
