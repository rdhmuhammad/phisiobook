package controller

import (
	"base-be-golang/pkg/chat_io"
	"base-be-golang/pkg/dto"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type ChatController struct {
	hub *chat_io.Hub
}

func NewChatController(hub *chat_io.Hub) ChatController {
	return ChatController{
		hub: hub,
	}
}

func (ctrl ChatController) EnterChat(c *gin.Context) {
	session := c.Query("session")
	if session == "" {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorResponse(fmt.Errorf("session is empty")))
		return
	}

	split := strings.Split(session, "x")
	if len(split) != 2 {
		c.JSON(http.StatusBadRequest, dto.DefaultErrorResponse(fmt.Errorf("session is invalid")))
		return
	}

	err := ctrl.hub.EnterRoom(c, split[0], split[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.DefaultErrorResponse(fmt.Errorf("session is invalid")))
		return
	}
}

func (ctrl ChatController) Route(route *gin.RouterGroup) {
	chatRoute := route.Group("/chat")
	chatRoute.GET("/enter", ctrl.EnterChat)
}
