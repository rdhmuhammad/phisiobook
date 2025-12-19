package main

import (
	"base-be-golang/pkg/api"
	"flag"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"log"
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

	err = app.Start()
	if err != nil {
		panic(err)
	}

}
