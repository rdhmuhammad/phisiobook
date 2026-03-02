package api

import (
	"github.com/rdhmuhammad/phisiobook/pkg/mongodb"
	"github.com/rdhmuhammad/phisiobook/shared/base"

	"gorm.io/gorm"
)

type Conns struct {
	Db      *gorm.DB
	MongoDb *mongodb.Conn
}

func (a *Api) RegisterSocket(r func(conns Conns, port base.Port, sct base.BaseSocket) []Namespace) {
	namespaces := r(Conns{
		Db:      a.db,
		MongoDb: a.mongoConn,
	},
		base.NewPort(a.db, a.cache, a.minioStr, a.reZero),
		base.NewBaseSocket(a.cache, a.db),
	)

	for _, namespace := range namespaces {
		a.namespaces = append(a.namespaces, namespace)
	}

}

func (a *Api) Register(r func(conns Conns, port base.Port, controller base.BaseController) []Router) {
	routers := r(Conns{
		Db:      a.db,
		MongoDb: a.mongoConn,
	},
		base.NewPort(a.db, a.cache, a.minioStr, a.reZero),
		base.NewBaseController(a.db, a.cache),
	)

	for _, router := range routers {
		a.routers = append(a.routers, router)
	}
}
