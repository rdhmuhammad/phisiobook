package main

import (
	"github.com/gin-gonic/gin"
	"github.com/zishang520/socket.io/servers/socket/v3"
	"log"
	"net/http"
)

func main() {
	api := gin.Default()
	server := socket.NewServer(nil, nil)

	server.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		log.Println(client)
		client
		// Handle events
		client.On("message", func(data ...any) {
			log.Println(data)
			// Echo the received message
			client.Emit("message", data...)
		})
	})

	api.Handle(http.MethodGet, "/socket.io/", gin.WrapH(server.ServeHandler(nil)))
	api.Run(":8080")
}
