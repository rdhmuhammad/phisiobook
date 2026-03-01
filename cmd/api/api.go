package main

import (
	"base-be-golang/internal/adapter/controller"
	"base-be-golang/internal/adapter/socket"
	"base-be-golang/internal/core/port"
	"base-be-golang/pkg/api"
	"flag"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	var envFile string
	flag.StringVar(&envFile, "env", ".env.stag", "Provide env file path")
	flag.Parse()

	err := godotenv.Load(envFile)
	if err != nil {
		log.Println(err)
		panic(err)

	}

	app := api.Default()

	// ========================= REGISTER CONTROLLER =========================
	app.Register(func(conn api.Conns, port port.Port, ctrl controller.BaseController) []api.Router {
		return []api.Router{
			controller.NewHomepageController(conn.Db, ctrl, port),
			controller.NewAuthController(conn.Db, ctrl, port),
			controller.NewHealthController(conn.Db, ctrl, port),
			controller.NewUserManagementController(conn.Db, ctrl, port),
			controller.NewChatController(conn.MongoDb, ctrl, port),
		}
	})

	app.RegisterSocket(func(conns api.Conns, port port.Port, sct socket.BaseSocket) []api.Namespace {
		return []api.Namespace{
			socket.NewChatSocket(conns.MongoDb, sct, port),
		}
	})

	err = app.Start()
	if err != nil {
		panic(err)
	}

}
