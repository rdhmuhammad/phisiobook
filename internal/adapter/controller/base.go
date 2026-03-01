package controller

import (
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/dto"
	"base-be-golang/pkg/environment"
	"base-be-golang/pkg/localize"
	"base-be-golang/pkg/mapper"
	"base-be-golang/pkg/middleware"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zishang520/socket.io/servers/socket/v3"
	"gorm.io/gorm"
)

type AuthInterface interface {
	GetSessionFromContext(ctx context.Context) middleware.SessionDataUser
	SetSessionToContext(c *gin.Context, ctx context.Context) (context.Context, error)
	SignClaim(claim middleware.DefaultUserClaim) (string, error)
	Validate() gin.HandlerFunc
	Authorize(roles ...string) gin.HandlerFunc
	GetAuthDataFromContext(c *gin.Context) middleware.UserData
	GetSessionData(userId string) (middleware.SessionDataUser, error)
	GetSessionDataFromContext(c *gin.Context) (middleware.SessionDataUser, error)
}

type Idempotent interface {
	Idempotent(name string, paramKey string, lockTime time.Duration) gin.HandlerFunc
}

type MapperUtility interface {
	ErrorSocket(client *socket.Socket, err error)
	ErrorResponse(c *gin.Context, err error) bool
	NewResponse(c *gin.Context, res *dto.Response, err error)
}

type EnigmaUtility interface {
	Validate(c *gin.Context, payload any) map[string][]string
	BindAndValidate(c *gin.Context, payload any) map[string][]string
	BindQueryToFilter(c *gin.Context, payload interface{}) error
	BindQueryToFilterAndValidate(c *gin.Context, payload interface{}) map[string][]string
}

type Environment interface {
	CheckFlag(flag string) bool
	Get(key string) string
	GetInt(key string, defaultValue int) int
	GetUint(key string, defaultValue uint) uint
	GetFloat(key string, defaultValue float64) float64
	GetBranchID() uint
}

type BaseController struct {
	auth      AuthInterface
	enigma    EnigmaUtility
	mapper    MapperUtility
	idem      Idempotent
	env       Environment
	localizer localize.Language
}

func NewBaseController(cache cache.Cache, dbConn *gorm.DB) BaseController {
	return BaseController{
		localizer: localize.NewLanguage("resource/message"),
		auth:      middleware.NewAuth(dbConn),
		enigma:    middleware.NewEnigma(),
		mapper:    mapper.NewMapper(),
		env:       environment.NewEnvironment(),
		idem:      middleware.NewIdempotent(cache),
	}
}
