package items

import (
	"encoding/json"
	"fmt"
	"lesta-battleship/server-core/internal/game"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Action struct {
	Name string
	Args map[string]interface{}
}

func ParseScript(script string) ([]Action, error) {
	script = replaceSingleQuotes(script)
	var actions []Action
	err := json.Unmarshal([]byte(script), &actions)
	if err != nil {
		return nil, err
	}
	return actions, nil
}

func replaceSingleQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "\"")
}

func toFloat(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	}
	return 0, false
}

func evalExpr(expr string, params map[string]interface{}, prevRand float64) (interface{}, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return 0, nil
	}
	if strings.HasPrefix(expr, "$") {
		expr = expr[1:]
	}
	if n, err := strconv.ParseFloat(expr, 64); err == nil {
		return n, nil
	}
	if val, ok := params[expr]; ok {
		return val, nil
	}
	if expr == "FIELD_SIZE" {
		if val, ok := params["FIELD_SIZE"]; ok {
			return val, nil
		}
		return 10, nil
	}
	if strings.Contains(expr, "RAND") {
		randRe := regexp.MustCompile(`\{\s*\"RAND\"\s*:\s*\"None\"\s*\}`)
		if randRe.MatchString(expr) {
			val := float64(rand.Intn(10))
			return val, nil
		}
	}
	if strings.Contains(expr, "PREV_RAND") {
		prevRandRe := regexp.MustCompile(`\{\s*\"PREV_RAND\"\s*:\s*\"None\"\s*\}`)
		if prevRandRe.MatchString(expr) {
			return prevRand, nil
		}
	}
	if strings.Contains(expr, "+") || strings.Contains(expr, "-") {
		repl := func(s string) string {
			s = strings.TrimSpace(s)
			if strings.HasPrefix(s, "$") {
				s = s[1:]
			}
			if n, err := strconv.ParseFloat(s, 64); err == nil {
				return fmt.Sprintf("%f", n)
			}
			if val, ok := params[s]; ok {
				return fmt.Sprintf("%v", val)
			}
			if s == "FIELD_SIZE" {
				if val, ok := params["FIELD_SIZE"]; ok {
					return fmt.Sprintf("%v", val)
				}
				return "10"
			}
			return s
		}
		re := regexp.MustCompile(`([+-])`)
		parts := re.Split(expr, -1)
		ops := re.FindAllString(expr, -1)
		var nums []float64
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(p, "$") {
				p = p[1:]
			}
			if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
				v, _ := evalExpr(p, params, prevRand)
				if f, ok := toFloat(v); ok {
					nums = append(nums, f)
					continue
				}
			}
			n, err := strconv.ParseFloat(repl(p), 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse %s", p)
			}
			nums = append(nums, n)
		}
		if len(nums) == 0 {
			return nil, fmt.Errorf("cannot parse expr: %s", expr)
		}
		res := nums[0]
		for i, op := range ops {
			if op == "+" {
				res += nums[i+1]
			} else if op == "-" {
				res -= nums[i+1]
			}
		}
		return res, nil
	}
	if strings.HasPrefix(expr, "{") && strings.HasSuffix(expr, "}") {
		if strings.Contains(expr, "RAND") {
			val := float64(rand.Intn(10))
			return val, nil
		}
		if strings.Contains(expr, "PREV_RAND") {
			return prevRand, nil
		}
	}
	return expr, nil
}

func RunScript(script string, state *game.GameState, params map[string]interface{}) (string, error) {
	actions, err := ParseScript(script)
	if err != nil {
		return "", err
	}
	var lastResult string
	var prevRand float64
	rand.Seed(time.Now().UnixNano())

	for _, action := range actions {
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
		name := strings.ToUpper(action.Name)
		switch name {
		case "OPEN_CELL":
			x, okX := toFloat(action.Args["x"])
			y, okY := toFloat(action.Args["y"])
			if !okX || !okY {
				return "", fmt.Errorf("invalid args for open_cell")
			}
			lastResult = game.OpenCell(int(x), int(y), state)
		case "MAKE_SHOT":
			x, okX := toFloat(action.Args["x"])
			y, okY := toFloat(action.Args["y"])
			if !okX || !okY {
				return "", fmt.Errorf("invalid args for MAKE_SHOT")
			}
			cmd := &game.ShootCommand{Target: game.Coord{X: int(x), Y: int(y)}}
			err := cmd.Apply(state)
			if err != nil {
				return "", err
			}
			lastResult = "shot_done"
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
			if int(x) < 0 || int(x) >= 10 || int(y) < 0 || int(y) >= 10 {
				return "", fmt.Errorf("cell out of bounds")
			}
			state.Field[int(x)][int(y)] = cellStatus
			lastResult = "cell_status_set"
		case "SET_SHIP_COORDINATES":
			x, okX := toFloat(action.Args["x"])
			y, okY := toFloat(action.Args["y"])
			x2, okX2 := toFloat(action.Args["x2"])
			y2, okY2 := toFloat(action.Args["y2"])
			if !okX || !okY || !okX2 || !okY2 {
				return "", fmt.Errorf("invalid args for SET_SHIP_COORDINATES")
			}
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
			for _, coord := range ship.Coords {
				state.Field[coord.X][coord.Y] = game.Empty
			}
			for _, coord := range newCoords {
				if coord.X < 0 || coord.X >= 10 || coord.Y < 0 || coord.Y >= 10 {
					return "", fmt.Errorf("new ship position out of bounds")
				}
				state.Field[coord.X][coord.Y] = game.ShipCell
			}
			ship.Coords = newCoords
			state.Ships[shipID] = ship
			lastResult = "ship_coords_set"
		case "END_PLAYER_ACTION":
			lastResult = "end_action"
		default:
			return "", fmt.Errorf("unknown action: %s", action.Name)
		}
	}
	return lastResult, nil
}
