package ws

import (
	"encoding/json"
	"lesta-battleship/server-core/internal/game"
	"lesta-battleship/server-core/internal/match"
	"lesta-battleship/server-core/internal/transaction"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func WebSocketHandler(c *gin.Context) {
	roomID := c.Query("room_id")
	playerID := c.Query("player_id")

	rawRoom, ok := match.Rooms.Load(roomID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("[WS] Upgrade error:", err)
		return
	}

	room := rawRoom.(*match.GameRoom)

	var player *match.PlayerConn
	if room.Player1.ID == playerID {
		player = room.Player1
		room.Player1.Conn = conn
	} else if room.Player2.ID == playerID {
		player = room.Player2
		room.Player2.Conn = conn
	} else {
		log.Println("[WS] Invalid playerID:", playerID)
		conn.Close()
		return
	}

	log.Printf("[WS] Player %s connected to room %s\n", playerID, roomID)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("[WS] Read error:", err)
			break
		}

		var input struct {
			Event string    `json:"event"`
			Ship  game.Ship `json:"ship"`
			X     int       `json:"x"`
			Y     int       `json:"y"`
		}
		_ = json.Unmarshal(msg, &input)

		log.Printf("[WS] Event received from %s: %s\n", playerID, input.Event)

		switch input.Event {

		case "place_ship":
			room.Mutex.Lock()

			if len(player.State.Ships) >= 10 {
				room.Mutex.Unlock()
				send(conn, "place_ship_error", "maximum 10 ships allowed")
				continue
			}

			// Command with Ship from client (without ID)
			cmd := &game.PlaceShipCommand{Ship: input.Ship}
			tx := transaction.NewTransaction()
			tx.Add(cmd)

			err := tx.Execute(player.State)

			// Get the auto-generated ID from GameState after placement
			shipID := cmd.Ship.ID
			shipType := cmd.Ship.Type

			room.Mutex.Unlock()

			if err != nil {
				send(conn, "place_ship_error", err.Error())
			} else {
				send(conn, "ship_placed", map[string]any{
					"ship_id":   shipID,
					"ship_type": shipType,
				})
			}

		case "ready":
			room.Mutex.Lock()

			// shipCount := len(player.State.Ships)
			// if shipCount < 10 {
			// 	send(conn, "not_enough_ships", fmt.Sprintf("you need to place %d more ships", 10-shipCount))
			// 	room.Mutex.Unlock()
			// 	continue
			// }

			player.Ready = true
			allReady := room.Player1.Ready && room.Player2.Ready
			shouldStart := false

			if allReady && room.Status == "waiting" {
				room.Status = "playing"
				room.Turn = room.Player1.ID
				shouldStart = true
			}
			room.Mutex.Unlock()

			send(conn, "ready_confirmed", gin.H{"all_ready": allReady})

			if shouldStart {
				log.Printf("[WS] Game started in room %s. First turn: %s\n", roomID, room.Turn)
				broadcast(room, "game_start", gin.H{"first_turn": room.Turn})
			}

		case "remove_ship":
			room.Mutex.Lock()

			if player.Ready {
				room.Mutex.Unlock()
				send(conn, "remove_ship_error", "you cannot remove ship after ready")
				break
			}

			shipID := input.Ship.ID
			if shipID == "" {
				room.Mutex.Unlock()
				send(conn, "remove_ship_error", "missing ship ID")
				break
			}

			cmd := &game.RemoveShipCommand{ShipID: shipID}
			tx := transaction.NewTransaction()
			tx.Add(cmd)

			err := tx.Execute(player.State)
			room.Mutex.Unlock()

			if err != nil {
				send(conn, "remove_ship_error", err.Error())
			} else {
				send(conn, "ship_removed", map[string]string{"ship_id": shipID})
			}

		case "fire":
			room.Mutex.Lock()
			log.Printf("[FIRE] %s firing at (%d,%d)", playerID, input.X, input.Y)

			if room.Status != "playing" {
				send(conn, "error", "game not started")
				room.Mutex.Unlock()
				continue
			}

			if room.Turn != playerID {
				send(conn, "not_your_turn", nil)
				room.Mutex.Unlock()
				continue
			}

			var target *match.PlayerConn
			if room.Player1.ID == playerID {
				target = room.Player2
			} else {
				target = room.Player1
			}

			cmd := &game.ShootCommand{Target: game.Coord{X: input.X, Y: input.Y}}
			tx := transaction.NewTransaction()
			tx.Add(cmd)
			err := tx.Execute(target.State)
			if err != nil {
				log.Println("[FIRE] Error:", err)
				send(conn, "fire_error", err.Error())
				room.Mutex.Unlock()
				continue
			}

			// Check if all ships destroyed
			shipsLeft := 0
			for _, s := range target.State.Ships {
				for _, coord := range s.Coords {
					if target.State.Field[coord.X][coord.Y] == game.ShipCell {
						shipsLeft++
					}
				}
			}

			gameOver := shipsLeft == 0
			log.Printf("[FIRE] hit=%v gameOver=%v", target.State.Field[input.X][input.Y] == game.Hit, gameOver)

			broadcast(room, "fire_result", gin.H{
				"x":         input.X,
				"y":         input.Y,
				"hit":       target.State.Field[input.X][input.Y] == game.Hit,
				"next_turn": target.ID,
				"game_over": gameOver,
			})

			if gameOver {
				room.Status = "ended"
				room.WinnerID = playerID
				broadcast(room, "game_end", gin.H{"winner": playerID})
			} else {
				room.Turn = target.ID
			}
			room.Mutex.Unlock()
		}
	}
}

func send(conn *websocket.Conn, event string, data any) {
	err := conn.WriteJSON(map[string]any{
		"event": event,
		"data":  data,
	})
	if err != nil {
		log.Println("[WS] Send failed:", err)
	}
}

func broadcast(room *match.GameRoom, event string, data any) {
	msg := map[string]any{
		"event": event,
		"data":  data,
	}
	raw, _ := json.Marshal(msg)

	if room.Player1.Conn != nil {
		room.Player1.Conn.WriteMessage(websocket.TextMessage, raw)
	}
	if room.Player2.Conn != nil {
		room.Player2.Conn.WriteMessage(websocket.TextMessage, raw)
	}
}
