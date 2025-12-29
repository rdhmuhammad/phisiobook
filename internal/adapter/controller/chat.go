package controller

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/usecase/caching_chat"
	"base-be-golang/internal/core/usecase/chat"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/chat_io"
	"base-be-golang/pkg/dto"
	"base-be-golang/pkg/mongodb"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type ChatController struct {
	hub *chat_io.Hub
	BaseController
	cachedUc CachedChatUsecase
	chatUc   ChatSessionUsecase
}

func NewChatController(hub *chat_io.Hub, cacheConn cache.Cache, dbConn *gorm.DB, mongoConn mongodb.Conn) ChatController {
	return ChatController{
		hub:            hub,
		cachedUc:       caching_chat.NewUsecase(mongoConn),
		BaseController: NewBaseController(cacheConn, dbConn),
	}
}

type ChatSessionUsecase interface {
	EnterRoom(c *gin.Context, roleName string, roomCode string, userCode string) error
	CreateSession(ctx context.Context, roomCode string) (chat.CreateSessionResponse, error)
}

type CachedChatUsecase interface {
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

func (ctrl ChatController) CreateSession(c *gin.Context) {
	var roomCode = c.Param("roomCode")
	if roomCode == "" {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorInvalidDataWithMessage("room code is required"))
		return
	}

	result, err := ctrl.chatUc.CreateSession(c.Request.Context(), roomCode)
	ctrl.mapper.NewResponse(c, dto.NewSuccessResponse(result, ""), err)
}

func (ctrl ChatController) EnterChat(c *gin.Context) {
	roleName := c.Param("roleName")
	if roleName == "" {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorInvalidDataWithMessage("role name is required"))
		return
	}
	session := c.Query("session")
	if session == "" {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorResponse(fmt.Errorf("session is empty")))
		return
	}

	split := strings.Split(session, ".")
	if len(split) != 2 {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorResponse(fmt.Errorf("session is invalid")))
		return
	}

	err := ctrl.chatUc.EnterRoom(c, roleName, split[0], split[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.DefaultErrorResponse(fmt.Errorf("session is invalid")))
		return
	}
}

func (ctrl ChatController) Route(route *gin.RouterGroup) {
	chatRoute := route.Group("/chat")
	chatRoute.GET("/enter/:roleName", ctrl.EnterChat)

	restRouter := route.Group(
		"",
		ctrl.auth.Validate(),
		ctrl.auth.Authorize(constant.RoleIsUser, constant.RolesIsTerapis))
	restRouter.GET("/history", ctrl.GetCached)
	restRouter.GET("/create-session/:roomCode", ctrl.CreateSession)
}
