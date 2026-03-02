//go:generate mockery --all --inpackage --case snake

package mapper

import (
	"github.com/rdhmuhammad/phisiobook/pkg/dto"
	localerror2 "github.com/rdhmuhammad/phisiobook/pkg/localerror"
	"github.com/rdhmuhammad/phisiobook/pkg/localize"
	"github.com/rdhmuhammad/phisiobook/shared/payload"

	"io"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/zishang520/socket.io/servers/socket/v3"
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
	ErrorSocket(client *socket.Socket, err error)
	ErrorResponse(c *gin.Context, err error) bool
	NewResponse(c *gin.Context, res *payload.Response, err error)
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
