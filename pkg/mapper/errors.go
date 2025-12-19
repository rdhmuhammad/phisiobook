package mapper

import (
	"base-be-golang/internal/constant"
	"base-be-golang/internal/dto"
	localerror "base-be-golang/internal/localerror"
	localerror2 "base-be-golang/pkg/localerror"
	"base-be-golang/pkg/localize"
	"base-be-golang/pkg/middleware"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"net/http"
	"regexp"
	"strings"
)

// TODO: validation move to struct tags
func (m mapper) TranslateSQLErr(mySqlErr *mysql.MySQLError, methodName string) error {
	switch mySqlErr.Number {
	case DuplicateEntryCode:
		re := regexp.MustCompile(`for key '([^']+)`)
		msgs := re.FindStringSubmatch(mySqlErr.Message)
		if len(msgs) > 1 {
			switch msgs[1] {
			//case "kassir_users.uk_kassir_users_username":
			//	return localerror.InvalidDataError{Msg: ErrUserNameIsUsed.Error()}
			default:
				return mySqlErr
			}
		}
		return mySqlErr

	case DataTruncateCode:
		re := regexp.MustCompile(`for column '([^']+)`)
		msgs := re.FindStringSubmatch(mySqlErr.Message)
		if len(msgs) > 1 {
			switch msgs[1] {
			case "gender":
				//return localerror.InvalidDataError{
				//	Msg: ErrInvalidGender.Error(),
				//}
			default:
				return mySqlErr
			}
		}
		return fmt.Errorf("%s: %w", methodName, mySqlErr)

	case ForeignConstrainFailCode:
		re := regexp.MustCompile("CONSTRAINT `([^`]+)")
		msgs := re.FindStringSubmatch(mySqlErr.Message)
		if len(msgs) > 1 {
			switch msgs[1] {
			//case "fk_kassir_branches":
			//	return localerror.InvalidDataError{
			//		Msg: ErrBranchNotFound.Error(),
			//	}
			default:
				return mySqlErr
			}
		}

	default:
		return mySqlErr
	}

	return mySqlErr

}

func (receiver mapper) GetAuthDataFromContext(c *gin.Context) middleware.UserData {
	authDataStr, ok := c.Get("authData")
	if !ok {
		return middleware.UserData{}
	}
	authData := authDataStr.(middleware.UserData)
	return authData
}

func (m mapper) NewResponse(c *gin.Context, res *dto.Response, err error) {
	userData := m.GetAuthDataFromContext(c)
	if err != nil {
		if ok, invErr := m.IsInvalidDataError(err); ok {
			var templates = make([]localize.TemplatingData, 0)
			if invErr.DataToTemplated != nil {
				for key, val := range invErr.DataToTemplated {
					templates = append(templates, localize.TemplatingData{
						Name:  key,
						Value: val,
					})
				}
			}
			c.JSON(
				http.StatusBadRequest,
				dto.DefaultErrorInvalidDataWithMessage(m.localizer.GetLocalized(userData.Lang, err.Error(), templates...)),
			)
			return
		}
		if m.IsAccessControlError(err) {
			c.JSON(
				http.StatusUnauthorized,
				dto.DefaultErrorInvalidDataWithMessage(m.localizer.GetLocalized(userData.Lang, err.Error())),
			)
			return
		}
		middleware.CaptureError(c, err)
		fmt.Printf("ERROR: %s \n", err.Error())
		c.JSON(
			http.StatusInternalServerError,
			dto.DefaultErrorResponseWithMessage(m.localizer.GetLocalized(userData.Lang, constant.InternalError), err),
		)
		return
	}
	if res != nil {
		res.Message = m.localizer.GetLocalized(userData.Lang, res.Message)
		c.JSON(http.StatusOK, res)
		return
	}

	c.Status(http.StatusOK)
}

func (m mapper) IsInvalidDataError(err error) (bool, localerror2.InvalidDataError) {
	var invalidDataError localerror2.InvalidDataError
	if errors.As(err, &invalidDataError) {
		return true, invalidDataError
	}
	return false, invalidDataError
}

func (m mapper) IsAccessControlError(err error) bool {
	var invalidDataError localerror2.AccessControlError
	if errors.As(err, &invalidDataError) {
		return true
	}
	return false
}

func (m mapper) CompareSliceOfErr(errs []error, target error) bool {
	for _, err := range errs {
		if errors.Is(err, target) {
			return true
		}
		if m.ErrorIs(err, target) {
			return true
		}
	}

	return false
}

func (m mapper) ErrorIs(template error, targer error) bool {
	re := regexp.MustCompile(`\{[0-9]+}`)
	pattern := re.ReplaceAllString(template.Error(), ".+")

	match, err := regexp.MatchString(pattern+"$", targer.Error())
	if err != nil {

		return false
	}

	if match {
		return true
	}

	return false
}

func (m mapper) ReplaceLabelErr(template error, params ...string) error {
	customeErr := template.Error()
	for i, param := range params {
		customeErr = strings.Replace(
			customeErr,
			fmt.Sprintf("{%d}", i),
			param,
			-1,
		)
	}

	return fmt.Errorf(customeErr)
}
