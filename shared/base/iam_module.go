package base

import (
	"context"
	"github.com/rdhmuhammad/phisiobook/pkg/cache"
	"github.com/rdhmuhammad/phisiobook/pkg/logger"
	"github.com/rdhmuhammad/phisiobook/shared/payload"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	md "github.com/rdhmuhammad/github.com/rdhmuhammad/phisiobook/iam-module/pkg/middleware"

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
