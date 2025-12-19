package port

import (
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/clock"
	"base-be-golang/pkg/davinci"
	"base-be-golang/pkg/environment"
	"base-be-golang/pkg/mailing"
	"base-be-golang/pkg/mapper"
	"base-be-golang/pkg/middleware"
	"base-be-golang/pkg/miniostorage"
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"html/template"
	"io"
	"time"
)

type Port struct {
	Davinci Generator
	Env     Environment
	Clock   Clock
	Cache   cache.Cache
	Mailing Mailing
	Mapper  mapper.MapperUtility
	Auth    Auth
	Storage StorageService
	DB      *gorm.DB
}

func NewPort(dbConn *gorm.DB, cache cache.Cache, minioConn miniostorage.StorageMinio) Port {
	return Port{
		Davinci: davinci.DefaultDavinci(),
		Env:     environment.NewEnvironment(),
		Clock:   clock.Default(),
		Cache:   cache,
		Mailing: mailing.NewConfig(),
		Mapper:  mapper.NewMapper(),
		Auth:    middleware.NewAuth(dbConn),
		Storage: minioConn,
		DB:      dbConn,
	}
}

type StorageService interface {
	GetFile(ctx context.Context, fileName string) (*bytes.Buffer, error)
	StoreFile(ctx context.Context, fileName string, file io.Reader, fileSize int64) (minio.UploadInfo, error)
	DeleteFile(ctx context.Context, fileName string) error
	HealthCheck(ctx context.Context) error
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

type Auth interface {
	SignClaim(claim middleware.DefaultUserClaim) (string, error)
	GetUserLogin(ctx context.Context) middleware.UserData
}

type Mailing interface {
	NativeSendEmail(payload mailing.NativeSendEmailPayload) error
}

type Generator interface {
	GenerateHash(secretKey []byte, uniqueID string) (string, error)
	GenerateHashValue(session string, id string, i int) (string, error)
	DecryptMessage(key []byte, data string) (string, error)
	EncryptMessage(key, data []byte) (string, error)
	GenerateOTPCode(
		secret string,
		counter uint64,
	) (int, error)
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
	payload EmailBodyVerifyOTPPayload,
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
