package middleware

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/domain"
	"base-be-golang/internal/dto"
	"base-be-golang/internal/localerror"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/clock"
	"base-be-golang/pkg/davinci"
	"base-be-golang/pkg/db"
	"base-be-golang/pkg/environment"
	localerror2 "base-be-golang/pkg/localerror"
	"base-be-golang/pkg/localize"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"os"
	"strings"
	"time"
)

type Auth struct {
	cache         cache.Cache
	davinci       davinci.Engine
	mapper        mapperAuth
	env           environment.ENV
	localize      localize.Language
	clock         clock.CLOCK
	userAdminRepo db.GenericRepository[domain.UserAdmin]
	userRepo      db.GenericRepository[domain.User]
}

type ctxKey string

const AuthCodeContext = ctxKey("authCode")

func NewAuth(dbConn *gorm.DB) Auth {
	return Auth{
		cache:         cache.Default(),
		davinci:       davinci.DefaultDavinci(),
		clock:         clock.Default(),
		userRepo:      db.NewGenericeRepo(dbConn, domain.User{}),
		userAdminRepo: db.NewGenericeRepo(dbConn, domain.UserAdmin{}),
		mapper:        SharedMapper{},
		localize:      localize.NewLanguage("resource/message"),
		env:           environment.NewEnvironment(),
	}
}

type mapperAuth interface {
	GetBodyJSON(c *gin.Context) map[string]any
}

func (receiver Auth) SignClaim(claim DefaultUserClaim) (string, error) {
	method := jwt.SigningMethodHS256
	claim.ExpiresAt = jwt.NewNumericDate(receiver.clock.NowUTC().Add(time.Hour * time.Duration(receiver.env.GetInt("EXPIRED_TOKEN_JWT", 0))))
	token := &jwt.Token{
		Header: map[string]interface{}{
			"typ": "JWT",
			"alg": method.Alg(),
		},
		Claims: claim,
		Method: method,
	}
	secret := []byte(os.Getenv("SECRET"))
	tokenStr, err := token.SignedString(secret)
	if err != nil {

		return "", err
	}
	return tokenStr, nil
}

func (receiver Auth) GetSessionDataFromContext(c *gin.Context) (SessionDataUser, error) {
	authData := receiver.GetAuthDataFromContext(c)
	return receiver.GetSessionData(authData.UserId)
}

func (receiver Auth) GetSessionData(userId string) (SessionDataUser, error) {
	loginCacheKey := "LOGIN_KEY_"
	secretSession := receiver.env.Get("DEFAULT_SECRET_LOGIN_SESSION")
	sessionKey, err := receiver.davinci.GenerateHashValue(secretSession, userId, 10)
	if err != nil {

		return SessionDataUser{}, err
	}

	var sessionData SessionDataUser
	sessionStr, err := receiver.cache.Get(context.Background(), loginCacheKey+sessionKey)
	if err != nil {

		if errors.Is(redis.Nil, err) {
			return SessionDataUser{}, localerror2.InvalidDataError{Msg: ""}
		}
		return SessionDataUser{}, err
	}

	err = json.Unmarshal([]byte(sessionStr), &sessionData)
	if err != nil {

		return SessionDataUser{}, err
	}

	return sessionData, nil
}

func (receiver Auth) Authorize(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		var authData = UserData{}
		authDataStr, ok := c.Get("authData")
		if ok {
			authData = authDataStr.(UserData)
		}

		if slices.Contains(roles, authData.RoleName) {
			c.Next()
			return
		}

		response := dto.DefaultBadRequestResponse()
		response.Message = receiver.localize.GetLocalized(authData.Lang, constant.AccessNotAllowed)
		response.ResponseTime = fmt.Sprint(time.Since(start).Milliseconds(), " ms.")
		c.JSON(http.StatusUnauthorized, response)
		c.Abort()
	}
}

func (receiver Auth) setUserActivity(authData UserData) {
	if authData.RoleName == constant.RolesIsMobile {
		user, err := receiver.userRepo.FindOneByExpression(
			context.Background(),
			[]clause.Expression{db.Equal(authData.UserId, "auth_code")},
		)
		if err != nil {
			CaptureErrorUsecase(context.Background(), err)
			fmt.Println(err.Error())
			return
		}

		user.LastActive = sql.NullTime{Time: time.Now(), Valid: true}
		err = receiver.userRepo.UpdateSelectedCols(context.Background(), user, "last_active")
		if err != nil {
			CaptureErrorUsecase(context.Background(), err)
			fmt.Println(err.Error())
			return
		}
		return
	}

	user, err := receiver.userAdminRepo.FindOneByExpression(
		context.Background(),
		[]clause.Expression{db.Equal(authData.UserId, "auth_code")},
	)
	if err != nil {
		CaptureErrorUsecase(context.Background(), err)
		fmt.Println(err.Error())
		return
	}

	user.LastActive = sql.NullTime{Time: time.Now(), Valid: true}
	err = receiver.userAdminRepo.UpdateSelectedCols(context.Background(), user, "last_active")
	if err != nil {
		CaptureErrorUsecase(context.Background(), err)
		fmt.Println(err.Error())
		return
	}

}

func (receiver Auth) Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		tokenStr := strings.Replace(c.GetHeader("Authorization"), "Bearer ", "", -1)
		secret := os.Getenv("SECRET")
		token, err := receiver.parseToken(tokenStr, []byte(secret))
		if err != nil {

			response := dto.DefaultErrorResponseWithMessage(err.Error(), err)
			response.ResponseTime = fmt.Sprint(time.Since(start).Milliseconds(), " ms.")
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		authData, valid := receiver.getAuthData(token)

		userDataStruct := UserData{}
		err = userDataStruct.LoadFromMap(authData)
		if err != nil {
			if err != nil {
				response := dto.DefaultErrorResponse(err)
				response.Message = receiver.localize.GetLocalized(userDataStruct.Lang, constant.SessionExpired)
				response.ResponseTime = fmt.Sprint(time.Since(start).Milliseconds(), " ms.")
				c.JSON(http.StatusUnauthorized, response)
				c.Abort()
				return
			}
		}
		if valid {

			receiver.setUserActivity(userDataStruct)

			c.Set("authData", userDataStruct)
			olCtx := c.Request.Context()
			newCtx := context.WithValue(olCtx, AuthCodeContext, userDataStruct)
			c.Request = c.Request.WithContext(newCtx)
			c.Next()
			return
		}

		response := dto.DefaultErrorResponse(err)
		response.ResponseTime = fmt.Sprint(time.Since(start).Milliseconds(), " ms.")
		c.JSON(http.StatusUnauthorized, response)
		c.Abort()
	}
}

func (receiver Auth) parseToken(tokenStr string, secret []byte) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token format")
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
		return nil, localerror2.AccessControlError{Msg: constant.SessionExpired}
	}

	return nil, localerror2.AccessControlError{Msg: constant.SessionExpired}
}

func (receiver Auth) GetUserLogin(ctx context.Context) UserData {
	value := ctx.Value(AuthCodeContext)
	mapped, ok := value.(UserData)
	if !ok {
		fmt.Println("Cannot read struct with value ", value)
		return UserData{}
	}

	return mapped
}

func (receiver Auth) GetAuthDataFromContext(c *gin.Context) UserData {
	var authData = UserData{}
	authDataStr, ok := c.Get("authData")
	if !ok {
		return UserData{}
	}
	authDataMap := authDataStr.(map[string]interface{})
	err := authData.LoadFromMap(authDataMap)
	if err != nil {
		return UserData{}
	}
	if !ok {
		return UserData{}
	}
	return authData
}

func (receiver Auth) getAuthData(token *jwt.Token) (map[string]interface{}, bool) {
	claims, ok := token.Claims.(jwt.MapClaims)
	valid := ok && token.Valid
	if !ok {
		return nil, false
	}

	var authData map[string]interface{}

	if valid {
		authData = claims["userData"].(map[string]interface{})
	}

	return authData, valid
}

func (receiver Auth) GetSessionFromContext(ctx context.Context) SessionDataUser {
	var data SessionDataUser
	if oc, ok := ctx.Value(CtxKeySession).(SessionDataUser); ok {
		data = oc
	}

	return data
}

func (receiver Auth) SetSessionToContext(c *gin.Context, ctx context.Context) (context.Context, error) {
	fromContext, err := receiver.GetSessionDataFromContext(c)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, CtxKeySession, fromContext), nil
}
