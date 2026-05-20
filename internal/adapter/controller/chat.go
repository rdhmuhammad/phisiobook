package controller

import (
	"context"

	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/usecase/caching_chat"
	"github.com/rdhmuhammad/phisiobook/pkg/mongodb"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	dto "github.com/rdhmuhammad/phisiobook/shared/payload"
	"gorm.io/gorm"

	"net/http"

	"github.com/gin-gonic/gin"
)

type ChatController struct {
	base.BaseController
	cachedUc CachedChatUsecase
}

func NewChatController(dbConn *gorm.DB, mongoConn *mongodb.Conn, controller base.BaseController, port base.Port) ChatController {
	return ChatController{
		cachedUc:       caching_chat.NewUsecase(dbConn, mongoConn, port),
		BaseController: controller,
	}
}

type CachedChatUsecase interface {
	CacheChat(ctx context.Context, request caching_chat.CacheChatRequest)
	CacheRoom(ctx context.Context, request caching_chat.CacheRoomRequest, cancle context.CancelFunc) (caching_chat.CacheRoomResponse, error)
	GetCached(ctx context.Context, request caching_chat.GetCachedRequest) (dto.PaginationResponse[caching_chat.CachedChat], error)
	GetChatRoom(ctx context.Context) ([]caching_chat.ChatRoomListResponse, error)
}

func (ctrl ChatController) GetCached(c *gin.Context) {
	var request = caching_chat.GetCachedRequest{
		Filter: &dto.GetListQueryNoPeriod{},
	}
	if errs := ctrl.Enigma.BindQueryToFilter(c, &request); errs != nil {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorInvalidDataWithMessage(errs.Error()))
		return
	}

	request.Filter.SetIfEmpty()
	result, err := ctrl.cachedUc.GetCached(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(result, ""), err)
}

func (ctrl ChatController) GetChatRoom(c *gin.Context) {
	result, err := ctrl.cachedUc.GetChatRoom(c.Request.Context())
	ctrl.Mapper.NewResponse(c, dto.NewSuccessResponse(result, constant.GetChatList.String()), err)
}

func (ctrl ChatController) Route(route *gin.RouterGroup) {
	restRouter := route.Group(
		"/chat",
		ctrl.Security.Validate(),
		ctrl.Security.Authorize(constant.RoleIsUser, constant.RolesIsTerapis))
	restRouter.GET("/history", ctrl.GetCached)
	restRouter.GET("/room", ctrl.GetChatRoom)
}
