package port

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/domain"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/clock"
	"base-be-golang/pkg/davinci"
	"base-be-golang/pkg/db"
	"base-be-golang/pkg/environment"
	"base-be-golang/pkg/localerror"
	"base-be-golang/pkg/mailing"
	"base-be-golang/pkg/mapper"
	"base-be-golang/pkg/middleware"
	"base-be-golang/pkg/miniostorage"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/minio/minio-go/v7"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"html/template"
	"io"
	"time"
)

type Port struct {
	Davinci       Generator
	Env           Environment
	Clock         Clock
	Cache         cache.Cache
	Mailing       Mailing
	Mapper        mapper.MapperUtility
	Auth          Auth
	Storage       StorageService
	userRepo      db.GenericRepository[domain.User]
	userAdminRepo db.GenericRepository[domain.UserAdmin]
}

func NewPort(dbConn *gorm.DB, cache cache.Cache, minioConn miniostorage.StorageMinio) Port {
	return Port{
		Davinci:       davinci.DefaultDavinci(),
		Env:           environment.NewEnvironment(),
		Clock:         clock.Default(),
		userAdminRepo: db.NewGenericeRepo(dbConn, domain.UserAdmin{}),
		userRepo:      db.NewGenericeRepo(dbConn, domain.User{}),
		Cache:         cache,
		Mailing:       mailing.NewConfig(),
		Mapper:        mapper.NewMapper(),
		Auth:          middleware.NewAuth(dbConn),
		Storage:       minioConn,
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

func (uc Port) GetUserLogin(ctx context.Context) (domain.UserEntityInterface, error) {
	userLogin := uc.Auth.GetUserLogin(ctx)

	userStr, err := uc.Cache.Get(ctx, fmt.Sprintf("%s%s", constant.CacheKeyLogin, userLogin.UserId))
	if err != nil {
		middleware.CaptureErrorUsecase(ctx, err)
	}

	switch {
	case slices.Contains([]string{constant.RoleIsAdmin, constant.RolesIsTerapis}, userLogin.RoleName):
		var data domain.UserAdmin
		err = json.Unmarshal([]byte(userStr), &data)
		if err == nil {
			return &data, nil
		}
		data, err = uc.userAdminRepo.FindOneByExpression(ctx, []clause.Expression{
			db.Equal(userLogin.UserId, "auth_code"),
		})
		if err != nil {
			err = localerror.AccessNotAllowedUserNotFound(err)
			return &domain.UserAdmin{}, err
		}
		return &data, nil
	case userLogin.RoleName == constant.RoleIsUser:
		var data domain.User
		err = json.Unmarshal([]byte(userStr), &data)
		if err == nil {
			return &data, nil
		}
		data, err = uc.userRepo.FindOneByExpression(ctx, []clause.Expression{
			db.Equal(userLogin.UserId, "auth_code"),
		})
		if err != nil {
			err = localerror.AccessNotAllowedUserNotFound(err)
			return &domain.User{}, err
		}
		return &data, nil
	}

	return nil, fmt.Errorf("Role is not defined")
}

func (uc Port) RefreshUserCached(ctx context.Context, user domain.UserEntityInterface, userId string) {
	userBytes, err := json.Marshal(user)
	if err != nil {
		middleware.CaptureErrorUsecase(ctx, err)
	}
	err = uc.Cache.Set(ctx,
		fmt.Sprintf("%s%s", constant.CacheKeyLogin, userId),
		string(userBytes),
		time.Hour*time.Duration(uc.Env.GetInt("EXPIRED_TOKEN_JWT", 1)))
	if err != nil {
		middleware.CaptureErrorUsecase(ctx, err)
	}
}
