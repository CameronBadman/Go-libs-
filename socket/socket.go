// general purpose high performance web socket
package socket

import (
	"context"
	"net/http"

	"github.com/coder/websocket"
)

type Mesage struct {
	Type websocket.MessageType
	Data []byte
}

type socket struct {
	conn  *websocket.Conn
	ctx   context.Context
	read  chan Mesage
	write chan Mesage
}

func NewSocket(w http.ResponseWriter, r *http.Request, ctx context.Context) (*socket, error) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return nil, err
	}

	socket := &socket{
		conn:  conn,
		ctx:   ctx,
		read:  make(chan Mesage),
		write: make(chan Mesage),
	}

	return socket, nil
}

func Read(socket *socket) {
	for {
		select {
		case req := <-socket.Read:
		}
	}
}

func Write(socket *socket) {
	for {
		select {
		case req := <-socket.Write:
		}
	}
}
