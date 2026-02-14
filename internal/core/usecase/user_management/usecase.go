package user_management

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/core/domain"
	"base-be-golang/internal/core/port"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/db"
	"base-be-golang/pkg/dto"
	"base-be-golang/pkg/localerror"
	"base-be-golang/pkg/miniostorage"
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Usecase struct {
	port.Port
	userAdminRepo db.GenericRepository[domain.UserAdmin]
	userRepo      db.GenericRepository[domain.User]
	dbTrx         db.DBTransaction
}

func NewUsecase(gormDb *gorm.DB, cacheClient cache.Cache, minioClient miniostorage.StorageMinio) Usecase {
	return Usecase{
		Port:          port.NewPort(gormDb, cacheClient, minioClient),
		userAdminRepo: db.NewGenericeRepo(gormDb, domain.UserAdmin{}),
		userRepo:      db.NewGenericeRepo(gormDb, domain.User{}),
		dbTrx: db.NewDBTransaction(gormDb,
			db.NewGenericeRepoPointr(gormDb, domain.User{}),
		),
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
		return err
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
			err = localerror.NotFound(err, constant.UserNotFound)
			return err
		}
		encryptMessage = d.Password
	} else {
		encryptMessage, err = u.Davinci.EncryptMessage([]byte(u.Env.Get("ENCRYPT_MESSAGE_PASSWORD")), []byte(request.Password))
		if err != nil {
			return err
		}
	}

	user.Password = encryptMessage
	userLogin, err := u.GetUserLogin(ctx)
	if err != nil {
		return err
	}

	switch action {
	case ActionIsCreateUser:
		user.SetCreated(userLogin.GetEmail())
		_, err = u.userAdminRepo.Store(ctx, user)
		if err != nil {
			return err
		}
		break
	case ActionIsUpdateUser:
		user.SetUpdated(userLogin.GetEmail())
		err = u.userAdminRepo.Update(ctx, user)
		if err != nil {
			return err
		}

		u.RefreshUserCached(ctx, &user, user.AuthCode)
		break
	}

	return nil
}

func (u Usecase) GetDetail(ctx context.Context, id uint) (UserDetailItem, error) {

	// user admin repo
	if item, role, err := u.userAdminDetail(ctx, id); err != nil && !localerror.IsNotFoundStr(constant.UserNotFound, err) {
		return UserDetailItem{}, err
	} else if item != nil {
		detailItem := DefaultUserDetailItem(item)
		detailItem.Role = role.Name
		return detailItem, nil
	}

	return UserDetailItem{}, nil

}

func (u Usecase) userMobileDetail(ctx context.Context, id uint) (domain.UserEntityInterface, error) {
	userMobile, err := u.userRepo.FindOneByID(ctx, id)
	if err != nil {
		err = localerror.NotFound(err, constant.UserNotFound)
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
		err = localerror.NotFound(err, constant.UserNotFound)
		return nil, domain.MasterRole{}, err
	}
	role = userAdmin.Role
	return &userAdmin, role, nil
}

func (u Usecase) GetList(ctx context.Context, query dto.GetListQueryNoPeriod) (dto.PaginationResponse[domain.UserListItem], error) {
	users, total, err := u.userRepo.FindPagedByExpressionJoin(
		ctx,
		db.Query(),
		db.PaginationQuery{Page: query.Page, PerPage: query.PerPage},
		nil,
		nil,
		db.ExpressionOr,
	)
	if err != nil {
		return dto.PaginationResponse[domain.UserListItem]{}, err
	}

	var result = make([]domain.UserListItem, 0)
	for i, item := range users {
		result[i] = domain.UserListItem{
			ID:         item.ID,
			Email:      item.Email,
			Name:       item.FullName,
			RoleKey:    "",
			StatusKey:  "",
			LastActive: item.GetLastActive(),
		}
	}

	return dto.NewPagination(result, total, query.PerPage, query.Page), nil

}

func (u Usecase) DeleteAccount(ctx context.Context) error {
	userLogin, err := u.GetUserLogin(ctx)
	if err != nil {
		return err
	}

	u.dbTrx.Begin()
	defer func(err *error) {
		errTrx := u.dbTrx.End(*err)
		if errTrx != nil {
			println(errTrx)
		}
	}(&err)

	userMobile := userLogin.(*domain.User)
	err = u.Cache.Delete(ctx, fmt.Sprintf("%s%s", constant.CacheKeyLogin, userMobile.AuthCode))
	if err != nil {
		return err
	}

	userRepo := db.GetRepo(&u.dbTrx, domain.User{})
	err = userRepo.DeleteByID(ctx, userMobile.ID)
	if err != nil {
		return err
	}

	return nil
}
