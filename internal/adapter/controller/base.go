package controller

import (
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/dto"
	"base-be-golang/pkg/mapper"
	"base-be-golang/pkg/middleware"
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
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
	NewResponse(c *gin.Context, res *dto.Response, err error)
}

type EnigmaUtility interface {
	Validate(c *gin.Context, payload any) map[string][]string
	BindAndValidate(c *gin.Context, payload any) map[string][]string
	BindQueryToFilter(c *gin.Context, payload interface{}) error
	BindQueryToFilterAndValidate(c *gin.Context, payload interface{}) map[string][]string
}

type BaseController struct {
	auth   AuthInterface
	enigma EnigmaUtility
	mapper MapperUtility
	idem   Idempotent
}

func NewBaseController(cache cache.Cache, dbConn *gorm.DB) BaseController {
	return BaseController{
		auth:   middleware.NewAuth(dbConn),
		enigma: middleware.NewEnigma(),
		mapper: mapper.NewMapper(),
		idem:   middleware.NewIdempotent(cache),
	}
}
