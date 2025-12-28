package chat_io

import (
	"fmt"
	"sync"
)

type MutexRoom struct {
	mutex sync.Mutex
	rooms map[string]*Room
}

func (m *MutexRoom) DeleteOnce(room *Room) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	close(room.broadcast)
	delete(m.rooms, room.id)
}

func (m *MutexRoom) AddOnce(room *Room) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.rooms[room.id] = room
}

func (m *MutexRoom) Get(roomId string) (*Room, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if t, ok := m.rooms[roomId]; ok {
		return t, ok
	}
	return nil, false
}

type MutexActor struct {
	mutex  sync.Mutex
	actors map[string]*Actor
}

func (m *MutexActor) DeleteOnce(actor *Actor) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	close(actor.send)
	delete(m.actors, actor.id)
}

func (m *MutexActor) AddOnce(actor *Actor) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.actors[actor.id]; ok {
		return fmt.Errorf("actor %s already exists", actor.id)
	}
	m.actors[actor.id] = actor
	return nil
}

func (m *MutexActor) Get(actorId string) (*Actor, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if t, ok := m.actors[actorId]; ok {
		return t, ok
	}
	return nil, false
}

func (m *MutexActor) Size() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.actors)
}
