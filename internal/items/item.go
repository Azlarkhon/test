package items

import (
	"encoding/json"
	"fmt"
	"lesta-battleship/server-core/internal/game"
	"lesta-battleship/server-core/internal/transaction"
	"net/http"
	"strings"
)

type Item struct {
	Script string `json:"script"`
	ID     int    `json:"id"`
}

func GetItemsInfo() (map[int]*Item, error) {
	r, err := http.Get("http://37.9.53.107/items/")
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var items []Item
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		return nil, err
	}

	itemsMap := make(map[int]*Item)
	for i := range items {
		itemsMap[items[i].ID] = &items[i]
	}

	return itemsMap, nil
}

func UseItem(id int, state *game.GameState, itemsMap map[int]*Item, params map[string]interface{}) (string, error) {
	item, exists := itemsMap[id]
	if !exists {
		return "", fmt.Errorf("item with id %d not found", id)
	}

	// Создаем транзакцию
	tx := transaction.NewTransaction()

	// Парсим скрипт и создаем команды для транзакции
	actions, err := ParseScript(item.Script)
	if err != nil {
		return "", fmt.Errorf("failed to parse script: %w", err)
	}

	var prevRand float64
	for _, action := range actions {
		// Обрабатываем аргументы
		for k, v := range action.Args {
			if s, ok := v.(string); ok {
				val, err := evalExpr(s, params, prevRand)
				if err == nil {
					action.Args[k] = val
					if strings.Contains(strings.ToUpper(s), "RAND") {
						if f, ok := val.(float64); ok {
							prevRand = f
						}
					}
				}
			}
		}

		// Создаем команду в зависимости от типа действия
		name := strings.ToUpper(action.Name)
		switch name {
		case "OPEN_CELL":
			x, okX := toFloat(action.Args["x"])
			y, okY := toFloat(action.Args["y"])
			if !okX || !okY {
				return "", fmt.Errorf("invalid args for open_cell")
			}
			cmd := NewOpenCellCommand(int(x), int(y))
			tx.Add(cmd)

		case "MAKE_SHOT":
			x, okX := toFloat(action.Args["x"])
			y, okY := toFloat(action.Args["y"])
			if !okX || !okY {
				return "", fmt.Errorf("invalid args for MAKE_SHOT")
			}
			cmd := NewShootCommand(int(x), int(y))
			tx.Add(cmd)

		case "SET_CELL_STATUS":
			x, okX := toFloat(action.Args["x"])
			y, okY := toFloat(action.Args["y"])
			status, okS := action.Args["status"].(string)
			if !okX || !okY || !okS {
				return "", fmt.Errorf("invalid args for SET_CELL_STATUS")
			}
			var cellStatus game.CellState
			switch status {
			case "water":
				cellStatus = game.Empty
			case "ship":
				cellStatus = game.ShipCell
			case "shipwreck":
				cellStatus = game.Hit
			default:
				return "", fmt.Errorf("unknown cell status: %s", status)
			}
			cmd := NewSetCellStatusCommand(int(x), int(y), cellStatus)
			tx.Add(cmd)

		case "SET_SHIP_COORDINATES":
			x, okX := toFloat(action.Args["x"])
			y, okY := toFloat(action.Args["y"])
			x2, okX2 := toFloat(action.Args["x2"])
			y2, okY2 := toFloat(action.Args["y2"])
			if !okX || !okY || !okX2 || !okY2 {
				return "", fmt.Errorf("invalid args for SET_SHIP_COORDINATES")
			}

			// Находим корабль
			var shipID string
			for id, ship := range state.Ships {
				for _, coord := range ship.Coords {
					if coord.X == int(x) && coord.Y == int(y) {
						shipID = id
						break
					}
				}
				if shipID != "" {
					break
				}
			}
			if shipID == "" {
				return "", fmt.Errorf("ship not found at (%d,%d)", int(x), int(y))
			}

			ship := state.Ships[shipID]
			lenCoords := len(ship.Coords)
			newCoords := make([]game.Coord, lenCoords)
			for i := 0; i < lenCoords; i++ {
				if x2 == x {
					newCoords[i] = game.Coord{X: int(x2), Y: int(y2) + i}
				} else if y2 == y {
					newCoords[i] = game.Coord{X: int(x2) + i, Y: int(y2)}
				} else {
					return "", fmt.Errorf("invalid ship orientation")
				}
			}

			cmd := NewSetShipCoordinatesCommand(shipID, ship.Coords, newCoords)
			tx.Add(cmd)

		case "END_PLAYER_ACTION":
			// Это специальное действие, не требует команды
			continue

		default:
			return "", fmt.Errorf("unknown action: %s", action.Name)
		}
	}

	// Выполняем транзакцию
	err = tx.Execute(state)
	if err != nil {
		return "", fmt.Errorf("transaction failed: %w", err)
	}

	return "item_used_successfully", nil
}
