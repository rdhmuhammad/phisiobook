package chat_io

import "log"

type Room struct {
	broadcast chan Transporter

	id string

	actors *MutexActor

	register chan *Actor

	unregister chan *Actor

	hub *Hub
}

func NewRoom(roomId string, actor *Actor, hub *Hub) (*Room, error) {
	mutexActor := &MutexActor{actors: make(map[string]*Actor)}
	if err := mutexActor.AddOnce(actor); err != nil {
		return nil, err
	}
	return &Room{
		id:         roomId,
		broadcast:  make(chan Transporter),
		actors:     mutexActor,
		hub:        hub,
		register:   make(chan *Actor),
		unregister: make(chan *Actor),
	}, nil
}

func (r *Room) run() {
	for {
		select {
		case actor := <-r.register:
			if err := r.actors.AddOnce(actor); err != nil {
				log.Printf("error adding actor: %v", err)
			}
		case actor := <-r.unregister:
			r.actors.DeleteOnce(actor)
			if r.actors.Size() == 0 {
				log.Println("begin closing room")
				select {
				case r.hub.close <- r:
					return
				}
			}
		case msg := <-r.broadcast:
			actor, ok := r.actors.Get(msg.FromID)
			log.Println(actor, msg, ok)
			if !ok {
				continue
			}
			select {
			case actor.send <- msg:
				continue
			default:
				log.Printf("actor %s leaving \n", actor.id)
				r.actors.DeleteOnce(actor)
			}
		}
	}
}
