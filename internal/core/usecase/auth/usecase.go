package auth

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/domain"
	"base-be-golang/internal/core/port"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/db"
	"base-be-golang/pkg/localerror"
	"base-be-golang/pkg/mailing"
	"base-be-golang/pkg/middleware"
	"base-be-golang/pkg/miniostorage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"strconv"
	"strings"
	"time"
)

type Usecase struct {
	userRepo      db.GenericRepository[domain.User]
	userAdminRepo db.GenericRepository[domain.UserAdmin]
	port.Port
}

func NewUsecase(dbConn *gorm.DB, cache cache.Cache, conn miniostorage.StorageMinio) Usecase {
	return Usecase{
		userAdminRepo: db.NewGenericeRepo[domain.UserAdmin](dbConn, domain.UserAdmin{}),
		userRepo:      db.NewGenericeRepo[domain.User](dbConn, domain.User{}),
		Port:          port.NewPort(dbConn, cache, conn),
	}
}

func (u Usecase) Logout(ctx context.Context, role string) error {
	userSession := u.Auth.GetUserLogin(ctx)
	switch role {
	case constant.ContextDashboard:
		user, err := u.userAdminRepo.FindOneByExpression(ctx, []clause.Expression{db.Equal(userSession.UserId, "auth_code")})
		err = localerror.AccessNotAllowedUserNotFound(err)
		if err != nil {
			return err
		}

		user.AuthCode = "EXPIRED"
		err = u.userAdminRepo.UpdateSelectedCols(ctx, user, "auth_code")
		if err != nil {
			return err
		}
		break
	case constant.ContextMobile:
		user, err := u.userRepo.FindOneByExpression(ctx, []clause.Expression{db.Equal(userSession.UserId, "auth_code")})
		err = localerror.AccessNotAllowedUserNotFound(err)
		if err != nil {
			return err
		}

		user.AuthCode = "EXPIRED"
		err = u.userRepo.UpdateSelectedCols(ctx, user, "auth_code")
		if err != nil {
			return err
		}
		break
	}

	err := u.Cache.Delete(ctx, fmt.Sprintf("%s%s", constant.CacheKeyLogin, userSession.UserId))
	if err != nil {
		middleware.CaptureErrorUsecase(ctx, err)
	}

	return nil
}

func (u Usecase) Login(ctx context.Context, request LoginRequest) (LoginResponse, error) {

	var (
		user       domain.UserEntityInterface
		userMobile domain.User
		userAdmin  domain.UserAdmin
		err        error
	)
	switch request.Role {
	case constant.ContextMobile:
		data, err := u.userRepo.FindOneByExpression(ctx, []clause.Expression{
			db.Equal(request.Email, "email"),
			db.Equal(true, "is_verified"),
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return LoginResponse{}, localerror.InvalidData(constant.LoginPasswordMismatch)
			}
			return LoginResponse{}, err
		}
		user = &data
		userMobile = data

		if !data.GetIsVerified() {
			return LoginResponse{}, localerror.InvalidData(constant.LoginUnverified)
		}
		break
	case constant.ContextDashboard:
		data, err := u.userAdminRepo.FindOneByExpressionAndJoin(
			ctx,
			[]clause.Expression{db.Equal(request.Email, "email")},
			[]string{"Role"}, nil)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return LoginResponse{}, localerror.InvalidData(constant.LoginPasswordMismatch)
			}
			return LoginResponse{}, err
		}
		user = &data
		userAdmin = data
		break
	}

	if rawPas, err := u.Davinci.DecryptMessage([]byte(u.Env.Get("ENCRYPT_MESSAGE_PASSWORD")), user.GetPassword()); err != nil {
		return LoginResponse{}, err
	} else if rawPas != request.Password {
		return LoginResponse{}, localerror.InvalidData(constant.LoginPasswordMismatch)
	}

	issuedAt := jwt.NewNumericDate(u.Clock.Now(ctx))
	userReference, err := u.Davinci.GenerateHash([]byte(u.Env.Get("SECRET_USER_ID")), strconv.FormatUint(uint64(user.GetID()), 10))
	if err != nil {
		return LoginResponse{}, err
	}
	userDataToken := middleware.UserData{
		UserId: userReference,
		Email:  user.GetEmail(),
	}
	if request.Role == constant.ContextMobile {
		userCoord, err := u.Cache.Get(ctx, constant.CacheKeyUserCoordinate+userMobile.AuthCode)
		if err != nil && !errors.Is(err, redis.Nil) {
			middleware.CaptureErrorUsecase(ctx, err)
		}
		err = u.Cache.Delete(ctx, constant.CacheKeyUserCoordinate+userMobile.AuthCode)
		if err != nil {
			middleware.CaptureErrorUsecase(ctx, err)
		}

		lang := u.Env.Get("FALLBACK_LANG")
		if userMobile.Lang != "" {
			lang = userMobile.Lang
		} else {
			userMobile.Lang = lang
		}
		userDataToken.Lang = lang
		userDataToken.RoleName = constant.RolesIsTerapis
		userMobile.AuthCode = userReference
		err = u.userRepo.UpdateSelectedCols(ctx, userMobile, "auth_code", "lang")
		if err != nil {
			return LoginResponse{}, err
		}

		err = u.Cache.Set(ctx, constant.CacheKeyUserCoordinate+userReference, userCoord, time.Duration(0))
		if err != nil {
			middleware.CaptureErrorUsecase(ctx, err)
		}
	} else {
		userAdmin.AuthCode = userReference
		userDataToken.RoleName = userAdmin.Role.Name
		err = u.userAdminRepo.UpdateSelectedCols(ctx, userAdmin, "auth_code")
		if err != nil {
			return LoginResponse{}, err
		}
	}

	token, err := u.Auth.SignClaim(middleware.DefaultUserClaim{
		UserData: userDataToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   strconv.FormatUint(uint64(user.GetID()), 10),
			IssuedAt: issuedAt,
		},
	})
	if err != nil {
		return LoginResponse{}, err
	}

	userBytes, err := json.Marshal(user)
	err = u.Cache.Set(
		ctx,
		fmt.Sprintf("%s%s", constant.CacheKeyLogin, userReference),
		string(userBytes),
		time.Hour*time.Duration(u.Env.GetInt("EXPIRED_TOKEN_JWT", 1)),
	)
	if err != nil {
		middleware.CaptureErrorUsecase(ctx, err)
	}

	return LoginResponse{
		UserID:     user.GetID(),
		Email:      user.GetEmail(),
		Token:      token,
		IsVerified: user.GetIsVerified(),
	}, nil
}

func (u Usecase) Register(ctx context.Context, request RegisterRequest) (RegisterResponse, error) {
	exist, err := u.userRepo.IsExistCondition(
		ctx,
		db.Query(
			db.Equal(request.Email, "email"),
			db.Equal(true, "is_verified"),
		),
	)
	if err != nil {
		return RegisterResponse{}, err
	}
	if exist {
		return RegisterResponse{}, localerror.InvalidData(constant.RegisterEmailUsed)
	}

	encryptMessage, err := u.Davinci.EncryptMessage([]byte(u.Env.Get("ENCRYPT_MESSAGE_PASSWORD")), []byte(request.Password))
	if err != nil {
		return RegisterResponse{}, err
	}

	user := domain.User{
		Email:       request.Email,
		Password:    encryptMessage,
		FullName:    request.FullName,
		LangContent: u.Env.Get("FALLBACK_LANG_CONTENT"),
		IsVerified:  0,
	}
	user.SetCreated("system")
	user, err = u.userRepo.Store(ctx, user)
	if err != nil {
		return RegisterResponse{}, err
	}

	if u.Env.CheckFlag("EMAIL_VERIFICATION_OFF") {
		user.SetIsVerified(true)
		err = u.userRepo.UpdateSelectedCols(ctx, user, "is_verified")
		if err != nil {
			return RegisterResponse{}, err
		}
	} else {
		otpResult, err := u.GenerateAndSendOTP(
			ctx,
			SendOtpRequest{
				Name:   request.FullName,
				UserID: uint64(user.ID),
				Email:  request.Email,
			},
			false,
		)
		if err != nil {
			return RegisterResponse{}, err
		}

		user.OTPCode = otpResult.Otp
		err = u.userRepo.UpdateSelectedCols(ctx, user, "otp_code")
		if err != nil {
			return RegisterResponse{}, err
		}
	}

	return RegisterResponse{UserID: user.ID}, nil
}

func (u Usecase) VerifyAcc(ctx context.Context, request VerifyAccRequest) (VerifyAccResponse, error) {
	user, err := u.userRepo.FindOneByExpression(ctx, []clause.Expression{db.Equal(request.Email, "email")})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VerifyAccResponse{}, localerror.InvalidData(constant.EmailNotFound)
		}
		return VerifyAccResponse{}, err
	}

	isExpired, err := u.Cache.Get(ctx, constant.CacheKeyOTP+strconv.FormatInt(int64(request.Otp), 10))
	if err != nil {
		if errors.Is(redis.Nil, err) {
			return VerifyAccResponse{}, localerror.InvalidData(constant.VerifyOtpExpired)
		}
		return VerifyAccResponse{}, err
	}

	if parseBool, err := strconv.ParseBool(isExpired); err != nil {
		return VerifyAccResponse{}, err
	} else if !parseBool {
		return VerifyAccResponse{}, localerror.InvalidData(constant.VerifyOtpExpired)
	}
	if user.OTPCode != request.Otp {
		return VerifyAccResponse{IsVerified: false}, nil
	}

	user.SetIsVerified(true)
	err = u.userRepo.UpdateSelectedCols(ctx, user, "is_verified")
	if err != nil {
		return VerifyAccResponse{}, err
	}

	return VerifyAccResponse{IsVerified: true}, nil
}

func (u Usecase) ResendOTP(ctx context.Context, request SendOtpRequest) error {
	user, err := u.userRepo.FindOneByExpression(ctx, []clause.Expression{db.Equal(request.Email, "email")})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return localerror.InvalidData(constant.EmailNotFound)
		}
		return err
	}

	if user.GetIsVerified() {
		return nil
	}

	_, err = u.GenerateAndSendOTP(
		ctx,
		SendOtpRequest{
			Email:  request.Email,
			Name:   user.FullName,
			UserID: uint64(user.ID),
		}, true,
	)
	if err != nil {
		return err
	}

	return nil

}

func (u Usecase) GenerateAndSendOTP(
	ctx context.Context,
	emailPayload SendOtpRequest,
	regenerate bool,
) (
	result SendOtpResponse,
	err error,
) {
	var user domain.User
	if regenerate {
		user, err = u.userRepo.FindOneByID(ctx, emailPayload.UserID)
		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return SendOtpResponse{}, localerror.InvalidData(constant.EmailNotFound)
			}
			return SendOtpResponse{}, err
		}
		emailPayload.Email = user.Email
		emailPayload.Name = user.FullName

		if user.IsVerified == 1 {
			return SendOtpResponse{}, localerror.InvalidData(constant.UserAlreadyVerified)
		}
	}

	movingFactor := uint64(u.Clock.NowUnix() / 30)
	secret := u.Env.Get("HOTP_SECRET")
	otp, err := u.Davinci.GenerateOTPCode(secret, movingFactor)
	if err != nil {

		return SendOtpResponse{}, err
	}

	otpStr := strconv.Itoa(otp)

	var tmplData = port.EmailBodyVerifyOTPPayload{
		Name:       emailPayload.Name,
		OTPs:       strings.Split(otpStr, ""),
		VerifyPage: os.Getenv("FRONT_END_HOST") + "/register/verifikasi/" + strconv.Itoa(int(emailPayload.UserID)),
	}
	emailPayload.Content, err = u.GenerateEmailBodyVerifyOTP(ctx, tmplData)
	emailPayload.Subject = "Register User Verification"
	err = u.sendEmail(emailPayload)
	if err != nil {

		return SendOtpResponse{}, err
	}

	if regenerate {
		user.OTPCode = int32(otp)
		user.IsVerified = 0
		err := u.userRepo.UpdateSelectedCols(ctx, user, "otp_code", "is_verified")
		if err != nil {

			return SendOtpResponse{}, err
		}
	}

	err = u.Cache.Set(ctx, constant.CacheKeyOTP+otpStr, true, time.Minute*time.Duration(u.Env.GetUint("EXPARATION_OTP_TIME", 0)))
	if err != nil {
		return SendOtpResponse{}, err
	}

	return SendOtpResponse{
		Otp: int32(otp),
	}, nil
}

func (u Usecase) sendEmail(emailPayload SendOtpRequest) error {
	err := u.Mailing.NativeSendEmail(mailing.NativeSendEmailPayload{
		Host:     os.Getenv("SMPT_SERVER_HOST"),
		Port:     os.Getenv("SMPT_SERVER_PORT"),
		Subject:  emailPayload.Subject,
		Username: os.Getenv("SUPPORT_EMAIL"),
		Password: os.Getenv("SUPPORT_EMAIL_PASS"),
		SendTo:   emailPayload.Email,
		HtmlBody: emailPayload.Content,
	})
	if err != nil {
		return err
	}
	return nil
}
