//go:generate mockery --all --inpackage --case snake

package mapper

import (
	"base-be-golang/internal/dto"
	"base-be-golang/internal/localerror"
	localerror2 "base-be-golang/pkg/localerror"
	"base-be-golang/pkg/localize"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"io"
)

type mapper struct {
	localizer localize.Language
}

func NewMapper() MapperUtility {
	return mapper{
		localizer: localize.NewLanguage("resource/message"),
	}
}

type MapperUtility interface {
	NewResponse(c *gin.Context, res *dto.Response, err error)
	ReplaceLabelErr(template error, params ...string) error
	ErrorIs(template error, targer error) bool
	TranslateSQLErr(mySqlErr *mysql.MySQLError, methodName string) error
	IsInvalidDataError(err error) (bool, localerror2.InvalidDataError)
	IsAccessControlError(err error) bool
	FloatPrecision(val float64, precision int) float64
	Base64ToReader(base64String string) (io.Reader, int64, error)
	CompareSliceOfErr(errs []error, target error) bool
	ParseServiceDurationFormat(d string) (string, error)
	SortingByStructField(vals interface{}, fieldName string, sorting SortingDirection) interface{}
	UniqueByStructField(vals interface{}, fieldName string) interface{}
}
