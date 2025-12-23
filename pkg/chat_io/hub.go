package chat_io

type Hub struct {
	room map[string]*Room

	init chan *Room

	close chan *Room
}

func NewHub() *Hub {
	hub := &Hub{
		room: make(map[string]*Room),
	}
	go hub.run()

	return hub
}

func (h *Hub) EnterRoom(roomId string, actor *Actor) {
	if existing, ok := h.room[roomId]; ok {
		existing.register <- actor
		return
	}
	h.room[roomId] = NewRoom(actor)
	go h.room[roomId].run()
	go actor.read()
	
}

func (h *Hub) run() {
	for {
		select {
		case room := <-h.close:
			if existing, ok := h.room[room.id]; ok {
				existing.closeRun <- true
				close(existing.broadcast)
				delete(h.room, existing.id)
			}
		}
	}
}
