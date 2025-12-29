package chat_io

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

const (
	cachingWait = 10 * time.Second
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Actor struct {
	id   string
	send chan Transporter

	room *Room

	hub *Hub

	conn *websocket.Conn
}

// read from inbound mesage then send to designate actor
func (c *Actor) read() {
	log.Printf("start reader of %s", c.id)
	defer func() {
		log.Println("close reader")
		c.room.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		var message Transporter
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			return
		}

		c.caching(message)
		log.Println("message read: ", message)
		select {
		case c.room.broadcast <- message:
		}
	}
}

func (c Actor) caching(message Transporter) {
	if c.hub.caching != nil {
		newContext, cancle := context.WithDeadline(context.Background(), time.Now().Add(cachingWait))
		go c.hub.caching.Store(newContext, cancle, message)
	}
}

func (c *Actor) write() {
	log.Printf("start writter of %s", c.id)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Println("close writer")
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if c.w(ok, message) {
				return
			}
		case <-ticker.C:
			log.Println("ping ticker...")
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Actor) w(ok bool, message Transporter) bool {
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if !ok {
		// The hub closed the channel.
		c.conn.WriteMessage(websocket.CloseMessage, []byte{})
		return true
	}

	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		log.Println(err)
		return true
	}

	marshal, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return true
	}
	w.Write(marshal)

	//
	//// Add queued chat messages to the current websocket message.
	//n := len(c.send)
	//for i := 0; i < n; i++ {
	//	w.Write(newline)
	//	w.Write(<-c.send)
	//}

	if err := w.Close(); err != nil {
		return true
	}
	return false
}
