package api

import (
	"base-be-golang/internal/adapter/controller"
	"base-be-golang/internal/adapter/socket"
	"base-be-golang/internal/core/port"
	"base-be-golang/pkg/mongodb"

	"gorm.io/gorm"
)

type Conns struct {
	Db      *gorm.DB
	MongoDb *mongodb.Conn
}

func (a *Api) RegisterSocket(r func(conns Conns, port port.Port, sct socket.BaseSocket) []Namespace) {
	namespaces := r(Conns{
		Db:      a.db,
		MongoDb: a.mongoConn,
	},
		port.NewPort(a.db, a.cache, a.minioStr, a.reZero),
		socket.NewBaseSocket(a.cache, a.db),
	)

	for _, namespace := range namespaces {
		a.namespaces = append(a.namespaces, namespace)
	}

}

func (a *Api) Register(r func(conns Conns, port port.Port, controller controller.BaseController) []Router) {
	routers := r(Conns{
		Db:      a.db,
		MongoDb: a.mongoConn,
	},
		port.NewPort(a.db, a.cache, a.minioStr, a.reZero),
		controller.NewBaseController(a.cache, a.db),
	)

	for _, router := range routers {
		a.routers = append(a.routers, router)
	}
}
