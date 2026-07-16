package therapist

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/domain"
	"github.com/rdhmuhammad/phisiobook/pkg/db"
	"github.com/rdhmuhammad/phisiobook/pkg/localerror"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	"github.com/rdhmuhammad/phisiobook/shared/payload"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Usecase struct {
	base.Port
	dbConn                 *gorm.DB
	userAdminRepo          db.GenericRepository[domain.UserAdmin]
	therapistRepo          db.GenericRepository[domain.Therapist]
	therapistDocumentRepo  db.GenericRepository[domain.TherapistDocument]
	masterRoleRepo         db.GenericRepository[domain.MasterRole]
	onboardingApprovalRepo db.GenericRepository[domain.OnboardingApproval]
}

func NewUsecase(dbConn *gorm.DB, port base.Port) Usecase {
	return Usecase{
		Port:                   port,
		dbConn:                 dbConn,
		userAdminRepo:          db.NewGenericeRepo(dbConn, domain.UserAdmin{}),
		therapistRepo:          db.NewGenericeRepo(dbConn, domain.Therapist{}),
		therapistDocumentRepo:  db.NewGenericeRepo(dbConn, domain.TherapistDocument{}),
		masterRoleRepo:         db.NewGenericeRepo(dbConn, domain.MasterRole{}),
		onboardingApprovalRepo: db.NewGenericeRepo(dbConn, domain.OnboardingApproval{}),
	}
}

func (u Usecase) Register(ctx context.Context, request RegisterTherapistRequest) (response RegisterTherapistResponse, err error) {
	exist, err := u.userAdminRepo.IsExistCondition(ctx, db.Query(
		db.Equal(request.Email, "email"),
	))
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}
	if exist {
		return response, localerror.InvalidData(constant.TherapistEmailUsed.String())
	}

	terapisRole, err := u.masterRoleRepo.FindOneByExpression(ctx, []clause.Expression{
		db.Equal("TERAPIS", "name"),
	})
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	encryptMessage, err := u.Davinci.EncryptMessage([]byte(u.Env.Get("ENCRYPT_MESSAGE_PASSWORD")), []byte(request.Password))
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	decoded, size, err := decodeBase64(request.Profile)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}
	profileFileName := fmt.Sprintf("therapist/profile_%d", time.Now().UnixMilli())
	_, err = u.Storage.StoreFile(ctx, profileFileName, decoded, size)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	userAdminCode, err := u.Davinci.GenerateUniqueKeyWithPredicate(
		u.Env.Get("SECRET_USER_KEY"),
		request.Email,
		6,
		func(result string) (bool, error) {
			return u.userAdminRepo.IsExistCondition(ctx, db.Query(
				db.Equal(result, "code"),
			))
		},
	)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}
	userAdminCode = fmt.Sprintf("TRP-%s", strings.ToUpper(userAdminCode))

	therapistCode, err := u.Davinci.GenerateUniqueKeyWithPredicate(
		u.Env.Get("SECRET_USER_KEY"),
		request.Email,
		6,
		func(result string) (bool, error) {
			return u.therapistRepo.IsExistCondition(ctx, db.Query(
				db.Equal(result, "code"),
			))
		},
	)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}
	therapistCode = fmt.Sprintf("THR-%s", strings.ToUpper(therapistCode))

	trx := db.NewTransaction(u.dbConn)
	defer func() {
		if err != nil {
			trx.End(err)
		} else {
			trx.End(nil)
		}
	}()

	userAdmin := domain.UserAdmin{
		Code:     userAdminCode,
		FullName: request.FullName,
		Email:    request.Email,
		Phone:    request.Phone,
		RoleID:   terapisRole.ID,
		Password: encryptMessage,
		IsActive: 0,
	}
	userAdmin.SetCreated("system")

	userAdminRepo := db.GetRepo(trx, domain.UserAdmin{})
	userAdmin, err = userAdminRepo.Store(ctx, userAdmin)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	therapist := domain.Therapist{
		Code:           therapistCode,
		Profile:        profileFileName,
		Name:           request.FullName,
		IsVerified:     0,
		CityId:         db.NullBigint(0),
		ExperienceYear: 0,
		Price:          0,
		AuthID:         userAdmin.ID,
		TherapyID:      1,
	}
	therapist.SetCreated("system")

	therapistRepo := db.GetRepo(trx, domain.Therapist{})
	therapist, err = therapistRepo.Store(ctx, therapist)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	therapistDocument := domain.TherapistDocument{
		TherapistID: therapist.ID,
		BankName:    "",
		BankCode:    "",
		AccName:     "",
		AccNumber:   "",
	}
	therapistDocument.SetCreated("system")

	therapistDocumentRepo := db.GetRepo(trx, domain.TherapistDocument{})
	_, err = therapistDocumentRepo.Store(ctx, therapistDocument)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	return RegisterTherapistResponse{
		Code:  therapist.Code,
		Email: request.Email,
	}, nil
}

func (u Usecase) Onboarding(ctx context.Context, request OnboardingRequest) (response OnboardingResponse, err error) {
	var userLogin payload.SessionDataUser
	err = u.Security.GetSessionLogin(ctx, &userLogin)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	therapist, err := u.therapistRepo.FindOneByExpression(ctx, []clause.Expression{
		db.Equal(userLogin.ID, "auth_id"),
	})
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	doc, err := u.therapistDocumentRepo.FindOneByExpression(ctx, []clause.Expression{
		db.Equal(therapist.ID, "therapist_id"),
	})
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	ktpFileName := fmt.Sprintf("therapist/ktp_%d_%d", therapist.ID, time.Now().UnixMilli())
	_, err = u.Storage.StoreFile(ctx, ktpFileName, request.KtpFile.Reader, request.KtpFile.Size)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	sipFileName := fmt.Sprintf("therapist/sip_%d_%d", therapist.ID, time.Now().UnixMilli())
	_, err = u.Storage.StoreFile(ctx, sipFileName, request.SipFile.Reader, request.SipFile.Size)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	strFileName := fmt.Sprintf("therapist/str_%d_%d", therapist.ID, time.Now().UnixMilli())
	_, err = u.Storage.StoreFile(ctx, strFileName, request.StrFile.Reader, request.StrFile.Size)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	ijazahFileName := fmt.Sprintf("therapist/ijazah_%d_%d", therapist.ID, time.Now().UnixMilli())
	_, err = u.Storage.StoreFile(ctx, ijazahFileName, request.IjazahFile.Reader, request.IjazahFile.Size)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	approvalCode, err := u.GenerateCode(ctx, "ONB-", func(ctx context.Context, code string) (bool, error) {
		return u.onboardingApprovalRepo.IsExistCondition(ctx, db.Query(
			db.Equal(code, "code"),
		))
	})
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	doc.KtpDoc = ktpFileName
	doc.StrDoc = strFileName
	doc.SipDoc = sipFileName
	doc.IjazahDoc = ijazahFileName
	doc.BankCode = request.BankCode
	doc.AccName = request.AccName
	doc.AccNumber = request.AccNumber
	doc.SetUpdated(userLogin.Email)

	trx := db.NewTransaction(u.dbConn)
	defer func() {
		if err != nil {
			trx.End(err)
		} else {
			trx.End(nil)
		}
	}()

	docRepo := db.GetRepo(trx, domain.TherapistDocument{})
	err = docRepo.UpdateSelectedCols(ctx, doc, "ktp_doc", "str_doc", "sip_doc", "ijazah_doc", "bank_code", "acc_name", "acc_number", "updated_by", "updated_at")
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	approval := domain.OnboardingApproval{
		Code:         approvalCode,
		TherapistID:  therapist.ID,
		LatestStatus: "PENDING",
		LatestReason: "",
		ApprovalByID: nil,
	}
	approval.SetCreated(userLogin.Email)

	approvalRepo := db.GetRepo(trx, domain.OnboardingApproval{})
	approval, err = approvalRepo.Store(ctx, approval)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	hist := domain.OnboardingApprovalHist{
		ApprovalID: approval.ID,
		Status:     "PENDING",
		Reason:     "",
	}
	hist.SetCreated(userLogin.Email)

	histRepo := db.GetRepo(trx, domain.OnboardingApprovalHist{})
	_, err = histRepo.Store(ctx, hist)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	return OnboardingResponse{
		Code:   therapist.Code,
		Status: "PENDING",
	}, nil
}

func (u Usecase) GetOnboardingList(ctx context.Context, query OnboardingListQuery) (response payload.PaginationResponse[OnboardingListItem], err error) {
	var conds []clause.Expression

	if query.Search != "" {
		conds = append(conds, db.Query(
			db.Search(query.Search, "latest_reason"),
			clause.Like{Column: "therapists.name", Value: "%" + query.Search + "%"},
			clause.Like{Column: "user_admins.full_name", Value: "%" + query.Search + "%"},
		)...)
	}

	if query.LatestStatus != "" {
		conds = append(conds, db.Equal(query.LatestStatus, "latest_status"))
	}

	approvals, total, err := u.onboardingApprovalRepo.FindPagedByExpressionJoin(
		ctx,
		conds,
		db.PaginationQuery{Page: query.Page, PerPage: query.PerPage},
		[]string{"Therapist", "ApprovalBy"},
		nil,
		db.ExpressionOr,
	)
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	items := make([]OnboardingListItem, len(approvals))
	for i, a := range approvals {
		approvalByName := ""
		if a.ApprovalBy != nil {
			approvalByName = a.ApprovalBy.FullName
		}
		items[i] = OnboardingListItem{
			Code:           a.Code,
			TherapistName:  a.Therapist.Name,
			LatestStatus:   a.LatestStatus,
			LatestReason:   a.LatestReason,
			CreatedAt:      a.CreatedAt,
			ApprovalByName: approvalByName,
		}
	}

	return payload.NewPagination(items, total, query.PerPage, query.Page), nil
}

func (u Usecase) GetOnboardingDetail(ctx context.Context, code string) (response OnboardingDetailResponse, err error) {
	approval, err := u.onboardingApprovalRepo.FindOneByExpressionAndJoin(
		ctx,
		db.Query(db.Equal(code, "onboarding_approvals.code")),
		[]string{"Therapist", "ApprovalBy"},
		nil,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response, localerror.InvalidData(constant.OnboardingNotFound.String())
		}
		return response, u.ErrHandler.ErrorReturn(err)
	}

	doc, err := u.therapistDocumentRepo.FindOneByExpression(ctx, []clause.Expression{
		db.Equal(approval.TherapistID, "therapist_id"),
	})
	if err != nil {
		return response, u.ErrHandler.ErrorReturn(err)
	}

	baseURL := u.Env.Get("BACKEND_URL")
	docURL := func(fileName string) string {
		if fileName == "" {
			return ""
		}
		return fmt.Sprintf("%s/api/v1/download?fileName=%s", baseURL, fileName)
	}

	approvalByName := ""
	if approval.ApprovalBy != nil {
		approvalByName = approval.ApprovalBy.FullName
	}

	therapistProfile := docURL(approval.Therapist.Profile)

	response = OnboardingDetailResponse{
		Code:             approval.Code,
		TherapistCode:    approval.Therapist.Code,
		TherapistName:    approval.Therapist.Name,
		TherapistProfile: therapistProfile,
		LatestStatus:     approval.LatestStatus,
		LatestReason:     approval.LatestReason,
		ApprovalByName:   approvalByName,
		CreatedAt:        approval.CreatedAt,
		UpdatedAt:        approval.UpdatedAt,
		KtpDoc:           docURL(doc.KtpDoc),
		StrDoc:           docURL(doc.StrDoc),
		SipDoc:           docURL(doc.SipDoc),
		IjazahDoc:        docURL(doc.IjazahDoc),
		BankName:         doc.BankName,
		BankCode:         doc.BankCode,
		AccName:          doc.AccName,
		AccNumber:        doc.AccNumber,
	}
	return response, nil
}

func (u Usecase) DeleteOnboarding(ctx context.Context, code string) (err error) {
	approval, err := u.onboardingApprovalRepo.FindOneByExpression(ctx, []clause.Expression{
		db.Equal(code, "code"),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return localerror.InvalidData(constant.OnboardingNotFound.String())
		}
		return u.ErrHandler.ErrorReturn(err)
	}

	trx := db.NewTransaction(u.dbConn)
	defer func() {
		if err != nil {
			trx.End(err)
		} else {
			trx.End(nil)
		}
	}()

	histRepo := db.GetRepo(trx, domain.OnboardingApprovalHist{})
	err = histRepo.DeleteByExpression(ctx, []clause.Expression{
		db.Equal(approval.ID, "approval_id"),
	})
	if err != nil {
		return u.ErrHandler.ErrorReturn(err)
	}

	approvalRepo := db.GetRepo(trx, domain.OnboardingApproval{})
	err = approvalRepo.DeleteByExpression(ctx, []clause.Expression{
		db.Equal(code, "code"),
	})
	if err != nil {
		return u.ErrHandler.ErrorReturn(err)
	}

	return nil
}

func decodeBase64(base64Str string) (io.Reader, int64, error) {
	data := base64Str
	if idx := strings.Index(data, ";base64,"); idx != -1 {
		data = data[idx+8:]
	}
	byts, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, 0, err
	}
	return bytes.NewReader(byts), int64(len(byts)), nil
}
