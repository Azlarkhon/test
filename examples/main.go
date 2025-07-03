package main

import (
	"lesta-battleship/server-core/internal/api"
	"lesta-battleship/server-core/internal/ws"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()
	r.POST("/start-match", api.StartMatch)
	r.GET("/ws", ws.WebSocketHandler)
	r.Run(":8080")
}
