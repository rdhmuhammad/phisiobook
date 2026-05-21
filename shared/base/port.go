package base

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"strconv"
	"time"

	md "iam_module/pkg/middleware"

	"github.com/minio/minio-go/v7"
	"github.com/rdhmuhammad/phisiobook/pkg/cache"
	"github.com/rdhmuhammad/phisiobook/pkg/clock"
	"github.com/rdhmuhammad/phisiobook/pkg/davinci"
	"github.com/rdhmuhammad/phisiobook/pkg/environment"
	"github.com/rdhmuhammad/phisiobook/pkg/localerror"
	"github.com/rdhmuhammad/phisiobook/pkg/localize"
	"github.com/rdhmuhammad/phisiobook/pkg/logger"
	"github.com/rdhmuhammad/phisiobook/pkg/mailing"
	"github.com/rdhmuhammad/phisiobook/pkg/mapper"
	"github.com/rdhmuhammad/phisiobook/pkg/middleware"
	"github.com/rdhmuhammad/phisiobook/pkg/miniostorage"
	"github.com/rdhmuhammad/phisiobook/shared/payload"
	"github.com/zishang520/socket.io/servers/socket/v3"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Port struct {
	Security   Security
	ErrHandler ErrHandler
	Cache      Cache
	Env        Environment
	Storage    StorageService
	Davinci    Generator
	Mailing    Mailing
	Clock      Clock
}

func NewPort(dbConn *gorm.DB, dbCache cache.DbClient, minioStr miniostorage.StorageMinio, zero *logger.ReZero) Port {
	return Port{
		Security:   NewAuth(dbConn, dbCache),
		ErrHandler: localerror.NewHandlerError(zero),
		Cache:      &dbCache,
		Storage:    minioStr,
		Env:        environment.NewEnvironment(),
		Davinci:    davinci.DefaultDavinci(),
		Mailing:    mailing.NewConfig(),
		Clock:      clock.Default(),
	}
}

type Security interface {
	SocketValidate(headerName string, idReqName string) socket.NamespaceMiddleware
	Validate() gin.HandlerFunc
	GetUserContext(ctx context.Context) payload.UserData
	Authorize(roles ...string) gin.HandlerFunc
	SetSession(ctx context.Context, user payload.SessionDataUser) error
	GetSession(ctx context.Context, authCode string, sessionData *payload.SessionDataUser) error
	GetSessionLogin(ctx context.Context, sessionData *payload.SessionDataUser) error
}

type Locale interface {
	GetLocalized(lang string, messageId string, templates ...localize.TemplatingData) string
}

type StorageService interface {
	GetFile(ctx context.Context, fileName string) (*bytes.Buffer, error)
	StoreFile(ctx context.Context, fileName string, file io.Reader, fileSize int64) (minio.UploadInfo, error)
	DeleteFile(ctx context.Context, fileName string) error
	HealthCheck(ctx context.Context) error
}

type ErrHandler interface {
	ErrorPrint(err error)
	DebugPrint(err string, v ...interface{})
	ErrorReturn(err error) error
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, keys ...string) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

type Mailing interface {
	NativeSendEmail(payload mailing.NativeSendEmailPayload) error
}

type Generator interface {
	GenerateUniqueKeyWithPredicate(
		secretKey string,
		uniqueID string,
		length int,
		isUnique davinci.UniquePredicate,
	) (string, error)
	GenerateUniqueKey(
		secretKey []byte,
		uniqueID string,
		length int,
	) (string, error)
	GenerateHash(secretKey []byte, uniqueID string) (string, error)
	GenerateHashValue(session string, id string, i int) (string, error)
	DecryptMessage(key []byte, data string) (string, error)
	EncryptMessage(key, data []byte) (string, error)
	GenerateOTPCode(
		secret string,
		counter uint64,
	) (int, error)
	GenerateCode(prefix string) string
}

type Environment interface {
	CheckFlag(flag string) bool
	Get(key string) string
	GetInt(key string, defaultValue int) int
	GetUint(key string, defaultValue uint) uint
	GetFloat(key string, defaultValue float64) float64
	GetBranchID() uint
}

type Clock interface {
	ParseWithTzFromCtx(ctx context.Context, value string, format string) time.Time
	Now(ctx context.Context) time.Time
	NowUTC() time.Time
	NowUnix() int64
	GetTimeZoneByName(name string) *time.Location
	SetTimezoneToContext(ctx context.Context, val string) context.Context
	GetTimezoneFromContext(ctx context.Context) *time.Location
}

type SendInBlueInterface interface {
	NativeSendEmail(ctx context.Context, payload mailing.NativeSendEmailPayload) error
}

func (uc Port) GenerateEmailBodyVerifyOTP(
	ctx context.Context,
	payload payload.EmailBodyVerifyOTPPayload,
) (string, error) {
	htmlPath := "./resource/mailing/verification-email.html"
	tmpl := template.Must(template.ParseFiles(htmlPath))
	outWriter := bytes.Buffer{}

	err := tmpl.Execute(&outWriter, payload)
	if err != nil {
		return "", err
	}

	return outWriter.String(), nil
}

func (uc Port) FormatRupiah(amount int) string {
	// Convert int to string
	str := strconv.FormatInt(int64(amount), 10)

	n := len(str)
	if n <= 3 {
		return "Rp. " + str + ",00"
	}

	var result string
	counter := 0

	// Loop from right to left
	for i := n - 1; i >= 0; i-- {
		result = string(str[i]) + result
		counter++

		if counter%3 == 0 && i != 0 {
			result = "." + result
		}
	}

	return "Rp. " + result + ",00"
}

func (uc Port) GenerateCode(ctx context.Context, prefix string, isExist func(ctx context.Context, code string) (bool, error)) (string, error) {
	code := uc.Davinci.GenerateCode(prefix)
	if exist, err := isExist(ctx, code); err != nil {
		return "", err
	} else if !exist {
		return code, nil
	}

	return uc.GenerateCode(ctx, prefix, isExist)
}

// ======================== BASE CONTROLLER ====================

type BaseController struct {
	Mapper   Mapper
	Locale   Locale
	Enigma   Validator
	Security Security
	Idem     Idempotent
}

func NewBaseController(db *gorm.DB, dbCache cache.DbClient) BaseController {
	return BaseController{
		Mapper:   mapper.NewMapper(),
		Idem:     middleware.NewIdempotent(dbCache),
		Enigma:   middleware.NewEnigma(),
		Locale:   localize.NewLanguage("resource/message"),
		Security: NewAuth(db, dbCache),
	}
}

type Mapper interface {
	ErrorSocket(client *socket.Socket, err error)
	ErrorResponse(c *gin.Context, err error) bool
	NewResponse(c *gin.Context, res *payload.Response, err error)
}

type Validator interface {
	BindQueryToFilterAndValidate(c *gin.Context, payload interface{}) map[string][]string
	BindAndValidate(c *gin.Context, payload any) map[string][]string
	BindQueryToFilter(c *gin.Context, payload interface{}) error
}

type Idempotent interface {
	Idempotent(name string, paramKey string, lockTime time.Duration) gin.HandlerFunc
}

// ======================== BASE CONTROLLER ====================

type BaseSocket struct {
	Security   Security
	Enigma     Validator
	Mapper     Mapper
	Clock      Clock
	ErrHandler ErrHandler
	Idem       Idempotent
	Locale     Locale
	Env        Environment
}

func NewBaseSocket(dbCache cache.DbClient, dbConn *gorm.DB, zero *logger.ReZero) BaseSocket {
	return BaseSocket{
		Security:   md.NewAuth(dbConn, dbCache),
		Enigma:     middleware.NewEnigma(),
		ErrHandler: localerror.NewHandlerError(zero),
		Clock:      clock.Default(),
		Locale:     localize.NewLanguage("resource/message"),
		Mapper:     mapper.NewMapper(),
		Env:        environment.NewEnvironment(),
		Idem:       middleware.NewIdempotent(dbCache),
	}
}
