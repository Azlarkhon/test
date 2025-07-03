package match

import (
	"lesta-battleship/server-core/internal/game"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type PlayerConn struct {
	ID    string
	Ready bool
	State *game.GameState
	Conn  *websocket.Conn
}

type GameRoom struct {
	RoomID    string
	Mode      string
	Player1   *PlayerConn
	Player2   *PlayerConn
	Status    string // waiting, ready, playing, ended
	Turn      string // player ID
	WinnerID  string
	Mutex     sync.Mutex
	CreatedAt time.Time
}

var Rooms sync.Map

func (p *PlayerConn) WriteMessage(msgType int, data []byte) error {
	if p.Conn != nil {
		return p.Conn.WriteMessage(msgType, data)
	}
	return nil
}
