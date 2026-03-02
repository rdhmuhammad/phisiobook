package main

import (
	"flag"

	iam "github.com/rdhmuhammad/phisiobook/iam_module/shared/adapter/controller"
	"github.com/rdhmuhammad/phisiobook/internal/adapter/controller"
	"github.com/rdhmuhammad/phisiobook/internal/adapter/socket"
	"github.com/rdhmuhammad/phisiobook/pkg/api"
	"github.com/rdhmuhammad/phisiobook/shared/base"

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
	app.Register(func(conn api.Conns, port base.Port, ctrl base.BaseController) []api.Router {
		return []api.Router{
			controller.NewHomepageController(conn.Db, ctrl, port),
			iam.NewAuthController(conn.Db, port, ctrl),
			controller.NewHealthController(conn.Db, ctrl, port),
			iam.NewUserManagementController(conn.Db, port, ctrl),
			controller.NewChatController(conn.MongoDb, ctrl, port),
		}
	})

	app.RegisterSocket(func(conns api.Conns, port base.Port, sct base.BaseSocket) []api.Namespace {
		return []api.Namespace{
			socket.NewChatSocket(conns.MongoDb, sct, port),
		}
	})

	err = app.Start()
	if err != nil {
		panic(err)
	}

}
