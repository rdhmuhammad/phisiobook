package base

import (
	"context"
	"os"
	"strconv"

	"github.com/rdhmuhammad/phisiobook/pkg/cache"
	"github.com/rdhmuhammad/phisiobook/pkg/logger"
	"github.com/rdhmuhammad/phisiobook/shared/payload"
	"github.com/zishang520/socket.io/servers/socket/v3"

	md "iam_module/pkg/middleware"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

func NewAuth(dbConn *gorm.DB, dbCache cache.DbClient) Security {
	if t, _ := strconv.ParseBool(os.Getenv("IAM_MODULE_OFF")); t {
		return EmptyAuth{}
	}
	return md.NewAuth(dbConn, dbCache)
}

// ======================= EMPTY AUTH ======================

// EmptyAuth implement if iam module is not used
type EmptyAuth struct {
}

func (e EmptyAuth) SocketValidate(headerName string) socket.NamespaceMiddleware {
	return func(s *socket.Socket, f func(*socket.ExtendedError)) {
		logger.Debug("using empty auth")
	}
}

func (e EmptyAuth) Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Debug("using empty auth")
	}
}

func (e EmptyAuth) GetUserContext(ctx context.Context) payload.UserData {
	return payload.UserData{}
}

func (e EmptyAuth) Authorize(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Debug("using empty auth")
	}
}

func (e EmptyAuth) SetSession(ctx context.Context, user payload.SessionDataUser) error {
	return nil
}

func (e EmptyAuth) GetSession(ctx context.Context, authCode string, sessionData *payload.SessionDataUser) error {
	return nil
}

func (e EmptyAuth) GetSessionLogin(ctx context.Context, sessionData *payload.SessionDataUser) error {
	return nil
}
