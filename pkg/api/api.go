package api

import (
	"os"

	"github.com/rdhmuhammad/phisiobook/pkg/cache"
	"github.com/rdhmuhammad/phisiobook/pkg/cio"
	"github.com/rdhmuhammad/phisiobook/pkg/logger"
	"github.com/rdhmuhammad/phisiobook/pkg/miniostorage"
	"github.com/rdhmuhammad/phisiobook/pkg/mongodb"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Api struct {
	server     *gin.Engine
	socket     *cio.IO
	db         *gorm.DB
	mongoConn  *mongodb.Conn
	cache      cache.DbClient
	minioStr   miniostorage.StorageMinio
	reZero     *logger.ReZero
	routers    []Router
	namespaces []Namespace
}

type Router interface {
	Route(handler *gin.RouterGroup)
}

type Namespace interface {
	OnSpace(nsfun cio.NSInitiate)
}

func (a *Api) Start() error {
	root := a.server.Group("/api/v1")
	for _, router := range a.routers {
		router.Route(root)
	}

	for _, namespace := range a.namespaces {
		namespace.OnSpace(a.socket.NewSpace)
	}

	port := os.Getenv("APP_PORT")
	err := a.server.Run("0.0.0.0:" + port)
	if err != nil {
		return err
	}

	return err
}
