package chat_io

type Room struct {
	broadcast chan Transporter

	id string

	actors map[string]*Actor

	register chan *Actor

	unregister chan *Actor

	closeRun chan bool

	hub *Hub
}

func NewRoom(actor *Actor) *Room {
	actors := make(map[string]*Actor)
	actors[actor.id] = actor
	return &Room{
		broadcast:  make(chan Transporter),
		actors:     actors,
		register:   make(chan *Actor),
		unregister: make(chan *Actor),
	}
}

func (r *Room) run() {
	for {
		select {
		case actor := <-r.register:
			r.actors[actor.id] = actor
		case actor := <-r.unregister:
			delete(r.actors, actor.id)
			close(actor.send)
			if len(r.actors) == 0 {
				r.hub.close <- r
			}
		case msg := <-r.broadcast:
			actor := r.actors[msg.toId]
			select {
			case actor.send <- msg:
			case <-r.closeRun:
				return
			default:
				close(actor.send)
				delete(r.actors, msg.toId)
			}
		}
	}
}
