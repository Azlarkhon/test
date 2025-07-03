package api

import (
	"lesta-battleship/server-core/internal/game"
	"lesta-battleship/server-core/internal/match"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func StartMatch(c *gin.Context) {
	var payload struct {
		RoomID  string `json:"room_id"`
		Player1 string `json:"player1"`
		Player2 string `json:"player2"`
		Mode    string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	room := &match.GameRoom{
		RoomID:    payload.RoomID,
		Mode:      payload.Mode,
		Player1:   &match.PlayerConn{ID: payload.Player1, State: game.NewGameState()},
		Player2:   &match.PlayerConn{ID: payload.Player2, State: game.NewGameState()},
		Status:    "waiting",
		CreatedAt: time.Now(),
	}
	match.Rooms.Store(payload.RoomID, room)
	c.JSON(http.StatusOK, gin.H{"status": "created"})
}
