package chat_io

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
)

type Hub struct {
	room *MutexRoom

	close chan *Room
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewHub() *Hub {
	hub := &Hub{
		room:  &MutexRoom{rooms: make(map[string]*Room)},
		close: make(chan *Room),
	}
	go hub.run()

	return hub
}

func (h *Hub) EnterRoom(ctx *gin.Context, roomId string, chatId string) error {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	var actor = &Actor{
		id:   chatId,
		send: make(chan Transporter),
		conn: conn,
	}

	if existing, ok := h.room.Get(roomId); ok {
		actor.room = existing
		existing.register <- actor
		go actor.read()
		go actor.write()
		return nil
	}
	newRoom, err := NewRoom(roomId, actor, h)
	if err != nil {
		return err
	}
	h.room.AddOnce(newRoom)
	actor.room = newRoom
	go newRoom.run()
	go actor.read()
	go actor.write()

	log.Printf("actors: %v\n", newRoom.actors)
	return nil
}

func (h *Hub) run() {
	log.Println("starting hub")
	for {
		select {
		case room := <-h.close:
			log.Printf("delete room: %v\n", room)
			h.room.DeleteOnce(room)
			log.Printf("room: %v", h.room)
		}
	}
}
