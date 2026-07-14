package registration

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rdhmuhammad/phisiobook/pkg/db"
	"github.com/rdhmuhammad/phisiobook/pkg/localerror"
	"github.com/rdhmuhammad/phisiobook/pkg/mailing"
	"github.com/rdhmuhammad/phisiobook/pkg/middleware"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	"github.com/rdhmuhammad/phisiobook/shared/payload"

	"iam_module/internal/core/constant"
	"iam_module/internal/core/domain"
	"iam_module/pkg/security"
	constant2 "iam_module/shared/constant"

	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Usecase struct {
	userRepo      db.GenericRepository[domain.User]
	userAdminRepo db.GenericRepository[domain.UserAdmin]
	auth          auth
	base.Port
}

type auth interface {
	GenerateSingleToken(claim security.SingleTokenClaim) (string, error)
}

func NewUsecase(dbConn *gorm.DB, port base.Port) Usecase {
	return Usecase{
		auth:          security.NewAuth(),
		Port:          port,
		userAdminRepo: db.NewGenericeRepo[domain.UserAdmin](dbConn, domain.UserAdmin{}),
		userRepo:      db.NewGenericeRepo[domain.User](dbConn, domain.User{}),
	}
}

func (u Usecase) Logout(ctx context.Context, role string) error {
	userSession := u.Security.GetUserContext(ctx)
	switch role {
	case constant.ContextDashboard:
		var user domain.UserAdmin
		err := setLogout(ctx, u.userRepo, &user)
		if err != nil {
			return u.ErrHandler.ErrorReturn(err)
		}
		break
	case constant.ContextMobile:
		var user domain.User
		err := setLogout(ctx, u.userRepo, &user)
		if err != nil {
			return u.ErrHandler.ErrorReturn(err)
		}
		break
	}

	err := u.Cache.Delete(ctx, constant.LoginCacheKey(userSession.UserId))
	if err != nil {
		u.ErrHandler.ErrorPrint(err)
		middleware.CaptureErrorUsecase(ctx, err)
	}

	return nil
}

type userLogout struct {
	ID       string `gorm:"column:id" json:"id"`
	AuthCode string `gorm:"column:auth_code" json:"auth_code"`
}

func setLogout[T schema.Tabler](ctx context.Context, repo db.GenericRepository[T], user domain.UserEntityInterface) error {
	var ul userLogout
	err := repo.FindOneByExpSelection(ctx,
		&ul,
		[]clause.Expression{db.Equal(user.GetAuthCode(), "auth_code")},
	)
	err = localerror.AccessNotAllowedUserNotFound(err)
	if err != nil {
		return err
	}

	user.SetAuthCode("EXPIRED")
	user.SetID(user.GetID())
	d := user.(T)
	err = repo.UpdateSelectedCols(ctx, d, "auth_code")
	if err != nil {
		return err
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
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return LoginResponse{}, localerror.InvalidData(constant2.LoginPasswordMismatch.String())
			}
			return LoginResponse{}, err
		}
		user = &data
		userMobile = data

		if !data.GetIsVerified() {
			return LoginResponse{Lang: userMobile.Lang}, localerror.InvalidData(constant2.LoginPasswordMismatch.String())
		}
		break
	case constant.ContextDashboard:
		data, err := u.userAdminRepo.FindOneByExpressionAndJoin(
			ctx,
			[]clause.Expression{db.Equal(request.Email, "email")},
			[]string{"Role"}, nil)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return LoginResponse{}, localerror.InvalidData(constant2.LoginPasswordMismatch.String())
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
		return LoginResponse{Lang: userMobile.Lang}, localerror.InvalidData(constant2.LoginPasswordMismatch.String())
	}

	userReference, err := u.Davinci.GenerateHash([]byte(u.Env.Get("SECRET_USER_ID")), strconv.FormatUint(uint64(user.GetID()), 10))
	if err != nil {
		return LoginResponse{}, err
	}

	userDataToken := security.UserData{
		UserId:   userReference,
		Email:    user.GetEmail(),
		Timezone: u.Env.Get("FALLBACK_TIMEZONE"),
	}

	if request.Timezone != "" {
		userDataToken.Timezone = request.Timezone
	}

	if request.Role == constant.ContextMobile {
		lang := u.Env.Get("FALLBACK_LANG")
		if userMobile.Lang != "" {
			lang = userMobile.Lang
		} else {
			userMobile.Lang = lang
		}
		userDataToken.Lang = lang
		userDataToken.RoleName = constant.RolesIsMobile
		userMobile.AuthCode = userReference
		err = u.userRepo.UpdateSelectedCols(ctx, userMobile, "auth_code", "lang")
		if err != nil {
			return LoginResponse{}, err
		}
	} else {
		userAdmin.AuthCode = userReference
		userDataToken.RoleName = userAdmin.Role.Name
		err = u.userAdminRepo.UpdateSelectedCols(ctx, userAdmin, "auth_code")
		if err != nil {
			return LoginResponse{}, err
		}
	}

	token, err := u.auth.GenerateSingleToken(security.SingleTokenClaim{
		UserData: userDataToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   strconv.FormatUint(uint64(user.GetID()), 10),
			IssuedAt: jwt.NewNumericDate(u.Clock.Now(ctx)),
		},
	})
	if err != nil {
		return LoginResponse{}, err
	}

	err = u.Security.SetSession(ctx, payload.SessionDataUser{
		ID:            user.GetID(),
		UserReference: userReference,
		RoleName:      userDataToken.RoleName,
		TimeZone:      userDataToken.Timezone,
		Lang:          userDataToken.Lang,
		Email:         user.GetEmail(),
		Name:          user.GetName(),
		IsVerified:    user.GetIsVerified(),
	})
	if err != nil {
		middleware.CaptureErrorUsecase(ctx, err)
	}

	return LoginResponse{
		Lang:       userMobile.Lang,
		Code:       user.GetCode(),
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
		return RegisterResponse{}, localerror.InvalidData(constant2.RegisterEmailUsed.String())
	}

	encryptMessage, err := u.Davinci.EncryptMessage([]byte(u.Env.Get("ENCRYPT_MESSAGE_PASSWORD")), []byte(request.Password))
	if err != nil {
		return RegisterResponse{}, err
	}

	code, err := u.Davinci.GenerateUniqueKeyWithPredicate(
		u.Env.Get("SECRET_USER_KEY"),
		request.Email,
		10,
		func(result string) (bool, error) {
			return u.userRepo.IsExistCondition(ctx, db.Query(
				db.Equal(result, "code"),
			))
		},
	)
	if err != nil {
		return RegisterResponse{}, err
	}

	user := domain.User{
		Email:      request.Email,
		Code:       code,
		Password:   encryptMessage,
		FullName:   request.FullName,
		IsVerified: 0,
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

	return RegisterResponse{Code: user.Code}, nil
}

func (u Usecase) VerifyAcc(ctx context.Context, request VerifyAccRequest) (VerifyAccResponse, error) {
	user, err := u.userRepo.FindOneByExpression(ctx, []clause.Expression{db.Equal(request.Email, "email")})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VerifyAccResponse{}, localerror.InvalidData(constant2.EmailNotFound.String())
		}
		return VerifyAccResponse{}, err
	}

	isExpired, err := u.Cache.Get(ctx, constant.CacheKeyOTP+strconv.FormatInt(int64(request.Otp), 10))
	if err != nil {
		if errors.Is(redis.Nil, err) {
			return VerifyAccResponse{}, localerror.InvalidData(constant2.VerifyOtpExpired.String())
		}
		return VerifyAccResponse{}, err
	}

	if parseBool, err := strconv.ParseBool(isExpired); err != nil {
		return VerifyAccResponse{}, err
	} else if !parseBool {
		return VerifyAccResponse{}, localerror.InvalidData(constant2.VerifyOtpExpired.String())
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
			return localerror.InvalidData(constant2.EmailNotFound.String())
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
				return SendOtpResponse{}, localerror.InvalidData(constant2.EmailNotFound.String())
			}
			return SendOtpResponse{}, err
		}
		emailPayload.Email = user.Email
		emailPayload.Name = user.FullName

		if user.IsVerified == 1 {
			return SendOtpResponse{}, localerror.InvalidData(constant2.UserAlreadyVerified.String())
		}
	}

	movingFactor := uint64(u.Clock.NowUnix() / 30)
	secret := u.Env.Get("HOTP_SECRET")
	otp, err := u.Davinci.GenerateOTPCode(secret, movingFactor)
	if err != nil {

		return SendOtpResponse{}, err
	}

	otpStr := strconv.Itoa(otp)

	var tmplData = payload.EmailBodyVerifyOTPPayload{
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
