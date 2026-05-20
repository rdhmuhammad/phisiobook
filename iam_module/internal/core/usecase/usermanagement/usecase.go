package user_management

import (
	"context"

	"github.com/rdhmuhammad/phisiobook/pkg/db"
	"github.com/rdhmuhammad/phisiobook/pkg/localerror"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	"github.com/rdhmuhammad/phisiobook/shared/payload"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"iam_module/internal/adapter/repository"
	"iam_module/internal/core/constant"
	"iam_module/internal/core/domain"
	constant2 "iam_module/shared/constant"
)

type Usecase struct {
	base.Port
	userAdminRepo db.GenericRepository[domain.UserAdmin]
	userRepo      db.GenericRepository[domain.User]
	userMainRepo  repository.UserRepo
}

func NewUsecase(gormDb *gorm.DB, port base.Port) Usecase {
	return Usecase{
		Port:          port,
		userMainRepo:  repository.NewUserRepo(gormDb),
		userAdminRepo: db.NewGenericeRepo(gormDb, domain.UserAdmin{}),
		userRepo:      db.NewGenericeRepo(gormDb, domain.User{}),
	}
}

const (
	ActionIsCreateUser = iota
	ActionIsUpdateUser
)

func (u Usecase) DeleteUser(ctx context.Context, id uint) error {
	return u.userAdminRepo.DeleteByID(ctx, id)
}

func (u Usecase) UpsertUser(ctx context.Context, request CreateUserRequest, action int) error {
	exist, err := u.userAdminRepo.IsExist(ctx, "email", request.Email)
	if err != nil {
		return u.ErrHandler.ErrorReturn(err)
	}
	if exist {
		return localerror.InvalidDataError{Msg: "Email already exists"}
	}

	var user = domain.UserAdmin{
		Email:    request.Email,
		FullName: request.FullName,
		RoleID:   request.RoleId,
	}
	var status bool
	if request.StatusKey != "" {
		status = request.StatusKey == domain.Active
	} else {
		status = true
	}
	user.SetIsVerified(status)

	var encryptMessage string
	if action == ActionIsUpdateUser && request.Password == "" {
		d, err := u.userAdminRepo.FindOneByID(ctx, request.ID)
		if err != nil {
			err = localerror.NotFound(err, constant2.UserNotFound.String())
			return u.ErrHandler.ErrorReturn(err)
		}
		encryptMessage = d.Password
	} else {
		encryptMessage, err = u.Davinci.EncryptMessage([]byte(u.Env.Get("ENCRYPT_MESSAGE_PASSWORD")), []byte(request.Password))
		if err != nil {
			return u.ErrHandler.ErrorReturn(err)
		}
	}

	user.Password = encryptMessage
	userLogin := u.Security.GetUserContext(ctx)
	switch action {
	case ActionIsCreateUser:
		user.SetCreated(userLogin.Email)
		_, err = u.userAdminRepo.Store(ctx, user)
		if err != nil {
			return u.ErrHandler.ErrorReturn(err)
		}
		break
	case ActionIsUpdateUser:
		user.SetUpdated(userLogin.Email)
		err = u.userAdminRepo.Update(ctx, user)
		if err != nil {
			return u.ErrHandler.ErrorReturn(err)
		}

		err = u.Security.SetSession(ctx, payload.SessionDataUser{
			UserReference: user.AuthCode,
			RoleName:      user.GetRoleName(),
			TimeZone:      userLogin.Timezone,
			Lang:          userLogin.Lang,
			PhoneNumber:   user.Phone,
			Email:         user.Email,
			Name:          user.FullName,
			IsVerified:    user.GetIsVerified(),
			ProfileImage:  "",
			LastActive:    user.LastActive.Time,
		})
		if err != nil {
			u.ErrHandler.ErrorPrint(err)
		}
		break
	}

	return nil
}

func (u Usecase) GetDetail(ctx context.Context, id uint) (UserDetailItem, error) {

	// user admin repo
	if item, role, err := u.userAdminDetail(ctx, id); err != nil && !localerror.IsNotFoundStr(constant2.UserNotFound.String(), err) {
		return UserDetailItem{}, err
	} else if item != nil {
		detailItem := DefaultUserDetailItem(item)
		detailItem.Role = role.Name
		return detailItem, nil
	}

	// user mobile repo
	if item, err := u.userMobileDetail(ctx, id); err != nil {
	} else {
		detailItem := DefaultUserDetailItem(item)
		detailItem.Role = constant.RolesIsMobile
		return detailItem, nil
	}

	return UserDetailItem{}, nil
}

func (u Usecase) userMobileDetail(ctx context.Context, id uint) (domain.UserEntityInterface, error) {
	userMobile, err := u.userRepo.FindOneByID(ctx, id)
	if err != nil {
		err = localerror.NotFound(err, constant2.UserNotFound.String())
		return nil, err
	}
	return &userMobile, nil
}

func (u Usecase) userAdminDetail(ctx context.Context, id uint) (domain.UserEntityInterface, domain.MasterRole, error) {
	var role domain.MasterRole
	userAdmin, err := u.userAdminRepo.FindOneByExpressionAndJoin(ctx,
		[]clause.Expression{db.Equal(id, "user_admins.id")},
		[]string{"Role"}, nil)
	if err != nil {
		err = localerror.NotFound(err, constant2.UserNotFound.String())
		return nil, domain.MasterRole{}, err
	}
	role = userAdmin.Role
	return &userAdmin, role, nil
}

func (u Usecase) GetList(ctx context.Context, query repository.UserListQuery) (payload.PaginationResponse[domain.UserListItem], error) {
	result, total, _, err := u.userMainRepo.UserDashboardList(ctx, query)
	if err != nil {
		return payload.PaginationResponse[domain.UserListItem]{}, err
	}

	return payload.NewPagination(result, total, query.Filter.PerPage, query.Filter.Page), nil

}

// ===================== USER MOBILE ======================

func (u Usecase) GetProfileMobile(ctx context.Context) (GetProfileResponse, error) {
	var userLogin payload.SessionDataUser
	err := u.Security.GetSessionLogin(ctx, &userLogin)
	if err != nil {
		return GetProfileResponse{}, err
	}

	return GetProfileResponse{
		ID:       userLogin.ID,
		FullName: userLogin.Name,
		Email:    userLogin.Email,
	}, err

}

func (u Usecase) DeleteAccount(ctx context.Context) error {
	var userLogin payload.SessionDataUser
	err := u.Security.GetSessionLogin(ctx, &userLogin)
	if err != nil {
		return err
	}

	err = u.Cache.Delete(ctx, constant.LoginCacheKey(userLogin.UserReference))
	if err != nil {
		return err
	}

	err = u.userRepo.DeleteByID(ctx, userLogin.ID)
	if err != nil {
		return err
	}

	return nil
}
