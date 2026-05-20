package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rdhmuhammad/phisiobook/pkg/cache"
	"github.com/rdhmuhammad/phisiobook/pkg/clock"
	"github.com/rdhmuhammad/phisiobook/pkg/db"
	"github.com/rdhmuhammad/phisiobook/pkg/environment"
	"github.com/rdhmuhammad/phisiobook/pkg/localerror"
	"github.com/rdhmuhammad/phisiobook/pkg/localize"
	"github.com/rdhmuhammad/phisiobook/pkg/logger"
	"github.com/rdhmuhammad/phisiobook/shared/payload"
	"github.com/zishang520/socket.io/servers/socket/v3"

	"iam_module/internal/core/constant"
	"iam_module/internal/core/domain"
	constant2 "iam_module/shared/constant"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Auth struct {
	cache         cache.DbClient
	env           environment.ENV
	localize      localize.Language
	clock         clock.CLOCK
	userRepo      db.GenericRepository[domain.User]
	userAdminRepo db.GenericRepository[domain.UserAdmin]
}

func NewAuth(dbConn *gorm.DB, dbCache cache.DbClient) Auth {
	return Auth{
		env:           environment.NewEnvironment(),
		cache:         dbCache,
		localize:      localize.NewLanguage("resource/message"),
		clock:         clock.CLOCK{},
		userRepo:      db.NewGenericeRepo(dbConn, domain.User{}),
		userAdminRepo: db.NewGenericeRepo(dbConn, domain.UserAdmin{}),
	}
}

func (receiver Auth) Authorize(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var authData = payload.UserData{}
		authDataStr, ok := c.Get(string(AuthCodeContext))
		if ok {
			authData = authDataStr.(payload.UserData)
		}

		if slices.Contains(roles, authData.RoleName) {
			c.Next()
			return
		}

		response := payload.DefaultBadRequestResponse()
		response.Message = receiver.localize.GetLocalized(authData.Lang, constant2.AccessNotAllowed.String())
		c.JSON(http.StatusUnauthorized, response)
		c.Abort()
	}
}

func (receiver Auth) SetSession(ctx context.Context, user payload.SessionDataUser) error {
	marshal, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return receiver.cache.Set(
		ctx,
		constant.LoginCacheKey(user.UserReference),
		string(marshal),
		time.Hour*time.Duration(receiver.env.GetInt("EXPIRED_TOKEN_JWT", 1)),
	)
}

func (receiver Auth) GetSessionLogin(ctx context.Context, sessionData *payload.SessionDataUser) error {
	userContext := receiver.GetUserContext(ctx)
	return receiver.GetSession(ctx, userContext.UserId, sessionData)
}

func (receiver Auth) GetSession(ctx context.Context, authCode string, sessionData *payload.SessionDataUser) error {
	sessionStr, err := receiver.cache.Get(ctx, constant.LoginCacheKey(authCode))
	if err != nil {
		if errors.Is(redis.Nil, err) {
			return localerror.AccessControlError{Msg: constant2.SessionExpired.String()}
		}
		return err
	}

	err = json.Unmarshal([]byte(sessionStr), sessionData)
	if err != nil {

		return err
	}

	return nil
}

func (receiver Auth) GetUserContext(ctx context.Context) payload.UserData {
	value := ctx.Value(AuthCodeContext)
	if value != nil {
		if marshal, err := json.Marshal(value.(payload.UserData)); err == nil {
			logger.Error(err)
		} else {
			logger.Debug("data catch from context => " + string(marshal))
		}
		return value.(payload.UserData)
	}

	logger.Debug("no data in context")
	return payload.UserData{}
}

func (receiver Auth) SocketValidate(headerName string) socket.NamespaceMiddleware {
	return func(client *socket.Socket, next func(*socket.ExtendedError)) {
		tokenAny := client.Handshake().Auth[headerName]
		token, ok := tokenAny.(string)

		if token == "" && receiver.env.CheckFlag("POSTMAN_FALLBACK_VALIDATION") {
			token = client.Handshake().Headers.Header().Get("Authorization")
			token = strings.TrimPrefix(token, "Bearer ")
			token = strings.TrimSpace(token)
			ok = true
		}

		if !ok || strings.TrimSpace(token) == "" {
			err := fmt.Errorf("missing token")
			logger.Debug(err.Error())
			next(socket.NewExtendedError("unauthorized",
				payload.DefaultErrorResponseWithMessage(err.Error(), err),
			))
			return
		}

		secret := os.Getenv("SECRET")
		_, err := receiver.parseToken(token, []byte(secret))
		if err != nil {
			logger.Debug(err.Error())
			next(socket.NewExtendedError("unauthorized",
				payload.DefaultErrorResponseWithMessage(err.Error(), err),
			))
			return
		}

		next(nil)
	}
}

/*
Validate user token, and attach token data to context
*/
func (receiver Auth) Validate() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenStr := strings.Replace(c.GetHeader("Authorization"), "Bearer ", "", -1)
		secret := os.Getenv("SECRET")
		token, err := receiver.parseToken(tokenStr, []byte(secret))
		if err != nil {
			response := payload.DefaultErrorResponseWithMessage(err.Error(), err)
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		authData, valid := receiver.getAuthData(token)
		userDataStruct := payload.UserData{}
		err = userDataStruct.LoadFromMap(authData)
		if err != nil {
			response := payload.DefaultErrorResponse(err)
			response.Message = receiver.localize.GetLocalized(userDataStruct.Lang, constant2.SessionExpired.String())
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		if valid {
			receiver.setUserActivity(userDataStruct)
			tz := time.UTC
			if userDataStruct.Timezone != "" {
				tz, err = time.LoadLocation(userDataStruct.Timezone)
				if err != nil {
					logger.Error(err)
				}
			}
			userDataStruct.Tz = tz
			c.Set(string(AuthCodeContext), userDataStruct)
			olCtx := c.Request.Context()
			newCtx := context.WithValue(olCtx, AuthCodeContext, userDataStruct)
			c.Request = c.Request.WithContext(newCtx)
			c.Next()
			return
		}

		response := payload.DefaultErrorResponse(err)
		c.JSON(http.StatusUnauthorized, response)
		c.Abort()
	}
}

func (receiver Auth) parseToken(tokenStr string, secret []byte) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Error(fmt.Errorf("invalid token format"))
			return nil, localerror.AccessControlError{Msg: constant2.AccessNotAllowed.String()}
		}
		return secret, nil
	})
	if err != nil {

		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if valid := claims.VerifyExpiresAt(receiver.clock.NowUTC().Unix(), true); valid {
			return token, nil
		}
		return nil, localerror.AccessControlError{Msg: constant2.SessionExpired.String()}
	}

	return nil, localerror.AccessControlError{Msg: constant2.SessionExpired.String()}
}

func (receiver Auth) getAuthData(token *jwt.Token) (map[string]interface{}, bool) {
	claims, ok := token.Claims.(jwt.MapClaims)
	valid := ok && token.Valid
	if !ok {
		return nil, false
	}
	return claims, valid
}

func (receiver Auth) setUserActivity(authData payload.UserData) {
	if authData.RoleName == constant.RolesIsMobile {
		var user domain.User
		setActivity(authData, receiver.userRepo, &user)
		return
	}
	var user domain.UserAdmin
	setActivity(authData, receiver.userAdminRepo, &user)
	return
}

type userSelect struct {
	id       uint   `gorm:"column:id" json:"id"`
	authCode string `gorm:"column:auth_code" json:"authCode"`
}

func setActivity[T schema.Tabler](authData payload.UserData, repo db.GenericRepository[T], user domain.UserEntityInterface) {
	var usec userSelect
	err := repo.FindOneByExpSelection(
		context.Background(),
		&usec,
		[]clause.Expression{db.Equal(authData.UserId, "auth_code")},
	)
	if err != nil {
		return
	}

	user.SetID(usec.id)
	user.SetLastActive(time.Now().UTC())
	d := user.(T)
	err = repo.UpdateSelectedCols(context.Background(), d, "last_active")
	if err != nil {
		return
	}
}
