package cio

import (
	"base-be-golang/pkg/localerror"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zishang520/socket.io/servers/socket/v3"
	"github.com/zishang520/socket.io/v3/pkg/types"
)

type IO struct {
	socket     *socket.Server
	clients    chan *socket.Socket
	errHandler localerror.HandleError
	ns         map[string]*NS
}

func New(server *gin.Engine) *IO {
	config := socket.DefaultServerOptions()
	config.SetPingInterval(300 * time.Millisecond)
	config.SetPingTimeout(200 * time.Millisecond)
	config.SetMaxHttpBufferSize(1000000)
	config.SetConnectTimeout(1000 * time.Millisecond)
	config.SetCors(&types.Cors{
		Origin:      "*",
		Credentials: true,
	})

	sc := socket.NewServer(server, config)

	return &IO{
		socket: sc,
	}
}

func (io *IO) NewSpace(name string, middleware types.EventListener) *NS {
	of := io.socket.Of(name, middleware)
	ns := newNS(io, of)

	if _, ok := io.ns[name]; ok {
		panic("duplicate namespace: " + name)
	}
	io.ns[name] = ns
	return ns
}

func (io *IO) GetSpace(name string) (*NS, bool) {
	ns, ok := io.ns[name]
	return ns, ok
}

// ================================ NameSpace ================================

type NS struct {
	Space        socket.Namespace
	useRoom      bool
	hub          *IO
	onConnect    func(n *NS, client *socket.Socket)
	onDisconnect func(n *NS, client *socket.Socket)
	onEvent      map[string]func(n *NS, client *socket.Socket, msg ...any)
}

type NSInitiate func(name string, md types.EventListener) *NS

func newNS(hub *IO, ns socket.Namespace) *NS {
	return &NS{
		Space: ns,
		hub:   hub,
	}
}

type NSListener func(io *NS, client *socket.Socket)

type MessagePayload interface {
	From(msg ...any)
}
type NSListenerMessage[T MessagePayload] func(io *NS, client *socket.Socket, msg T)

func (n *NS) UserRoom() *NS {
	n.useRoom = true
	return n
}

func (n *NS) Disconnect(md NSListener) *NS {
	n.onDisconnect = func(ns *NS, client *socket.Socket) {
		md(ns, client)
	}
	return n
}

func (n *NS) Connect(mds NSListener) *NS {
	n.onConnect = func(ns *NS, client *socket.Socket) {
		mds(ns, client)
	}
	return n
}

func (n *NS) Event(evname string, p MessagePayload, md NSListenerMessage[MessagePayload]) *NS {
	n.onEvent[evname] = func(ns *NS, client *socket.Socket, msg ...any) {
		p.From(msg...)
		md(ns, client, p)
	}
	return n
}

func (n *NS) Build() {
	n.Space.On("connection", func(a ...any) {
		client := a[0].(*socket.Socket)
		query := client.Handshake().Query

		if n.useRoom {
			roomId := query.Query().Get("roomId")
			client.Join(socket.Room(roomId))

			if n.onConnect != nil {
				n.onConnect(n, client)
			}

			for name, ev := range n.onEvent {
				client.On(name, func(msg ...any) {
					ev(n, client, msg)
				})
			}
			client.On("disconnect", func(any ...any) {
				if n.onDisconnect != nil {
					n.onDisconnect(n, client)
				}
			})
			return
		}

		if n.onConnect != nil {
			n.onConnect(n, client)
		}
		for name, ev := range n.onEvent {
			client.On(name, func(msg ...any) {
				ev(n, client, msg)
			})
		}
		client.On("disconnect", func(any ...any) {
			if n.onDisconnect != nil {
				n.onDisconnect(n, client)
			}
		})

	})
}
