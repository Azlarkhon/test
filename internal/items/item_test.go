package items

import (
	"fmt"
	"lesta-battleship/server-core/internal/game"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func printField(field [10][10]game.CellState) {
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			switch field[x][y] {
			case game.Empty:
				fmt.Print("~ ")
			case game.ShipCell:
				fmt.Print("S ")
			case game.Hit:
				fmt.Print("X ")
			case game.Miss:
				fmt.Print("o ")
			case game.Revealed:
				fmt.Print(". ")
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

func TestUseItem_FromSpec(t *testing.T) {
	// Крест Нахимова
	{
		state := game.NewGameState()
		item := Item{
			ID: 1,
			Script: `[
				{"Name":"open_cell","Args":{"x":"$x","y":"$y"}},
				{"Name":"open_cell","Args":{"x":"$x","y":"$y+1"}},
				{"Name":"open_cell","Args":{"x":"$x+1","y":"$y"}},
				{"Name":"open_cell","Args":{"x":"$x","y":"$y-1"}},
				{"Name":"open_cell","Args":{"x":"$x-1","y":"$y"}},
				{"Name":"END_PLAYER_ACTION","Args":{}}
			]`,
		}
		params := map[string]interface{}{"x": 5, "y": 5}
		_, err := UseItem(1, state, []Item{item}, params)
		if err != nil {
			t.Errorf("Крест Нахимова: %v", err)
		}
		t.Log("Крест Нахимова:")
		printField(state.Field)
	}

	// Ремонтный набор
	{
		state := game.NewGameState()
		item := Item{
			ID: 2,
			Script: `[
				{"Name":"SET_CELL_STATUS","Args":{"x":"$x","y":"$y","status":"ship"}},
				{"Name":"END_PLAYER_ACTION","Args":{}}
			]`,
		}
		params := map[string]interface{}{"x": 2, "y": 3, "status": "ship"}
		_, err := UseItem(2, state, []Item{item}, params)
		if err != nil {
			t.Errorf("Ремонтный набор: %v", err)
		}
		t.Log("Ремонтный набор:")
		printField(state.Field)
	}

	// Боевой приказ (без SWICH_CASE)
	{
		state := game.NewGameState()
		// Ставим корабль вручную
		ship := game.Ship{ID: "s1", Type: "destroyer", Coords: []game.Coord{{X: 1, Y: 1}, {X: 1, Y: 2}}}
		state.Ships[ship.ID] = ship
		state.Field[1][1] = game.ShipCell
		state.Field[1][2] = game.ShipCell
		item := Item{
			ID: 3,
			Script: `[
				{"Name":"SET_SHIP_COORDINATES","Args":{"x":"$x","y":"$y","x2":"$x2","y2":"$y2"}},
				{"Name":"END_PLAYER_ACTION","Args":{}},
				{"Name":"MAKE_SHOT","Args":{"x":"$x","y":"$y"}}
			]`,
		}
		params := map[string]interface{}{"x": 1, "y": 1, "x2": 3, "y2": 1}
		_, err := UseItem(3, state, []Item{item}, params)
		if err != nil {
			t.Errorf("Боевой приказ: %v", err)
		}
		t.Log("Боевой приказ:")
		printField(state.Field)
	}

	// Конь (SWICH_CASE не реализован, тестируем только одну ветку)
	{
		state := game.NewGameState()
		item := Item{
			ID: 4,
			Script: `[
				{"Name":"open_cell","Args":{"x":"$x","y":"$y"}},
				{"Name":"open_cell","Args":{"x":"$x","y":"$y+1"}},
				{"Name":"open_cell","Args":{"x":"$x","y":"$y+2"}},
				{"Name":"open_cell","Args":{"x":"$x-1","y":"$y+2"}},
				{"Name":"END_PLAYER_ACTION","Args":{}}
			]`,
		}
		params := map[string]interface{}{"x": 4, "y": 4}
		_, err := UseItem(4, state, []Item{item}, params)
		if err != nil {
			t.Errorf("Конь: %v", err)
		}
		t.Log("Конь:")
		printField(state.Field)
	}

	// Ладья (рандом по горизонтали)
	{
		state := game.NewGameState()
		item := Item{
			ID: 5,
			Script: `[
				{"Name":"open_cell","Args":{"x":"{\"RAND\":\"None\"}","y":"$y"}},
				{"Name":"open_cell","Args":{"x":"{\"RAND\":\"None\"}","y":"$y"}},
				{"Name":"open_cell","Args":{"x":"{\"RAND\":\"None\"}","y":"$y"}},
				{"Name":"open_cell","Args":{"x":"{\"RAND\":\"None\"}","y":"$y"}},
				{"Name":"open_cell","Args":{"x":"{\"RAND\":\"None\"}","y":"$y"}}
			]`,
		}
		params := map[string]interface{}{"y": 7}
		_, err := UseItem(5, state, []Item{item}, params)
		if err != nil {
			t.Errorf("Ладья: %v", err)
		}
		t.Log("Ладья:")
		printField(state.Field)
	}

	// Слон (рандом по диагонали, только одна ветка)
	{
		state := game.NewGameState()
		// Добавим корабли для наглядности
		state.Field[2][2] = game.ShipCell
		state.Field[3][3] = game.ShipCell
		state.Field[4][4] = game.ShipCell
		item := Item{
			ID: 6,
			Script: `[
				{"Name":"open_cell","Args":{"x":"{\"RAND\":\"None\"}-FIELD_SIZE+$x","y":"{\"PREV_RAND\":\"None\"}-FIELD_SIZE+$y"}},
				{"Name":"open_cell","Args":{"x":"{\"RAND\":\"None\"}-FIELD_SIZE+$x","y":"{\"PREV_RAND\":\"None\"}-FIELD_SIZE+$y"}},
				{"Name":"open_cell","Args":{"x":"{\"RAND\":\"None\"}-FIELD_SIZE+$x","y":"{\"PREV_RAND\":\"None\"}-FIELD_SIZE+$y"}},
				{"Name":"open_cell","Args":{"x":"{\"RAND\":\"None\"}-FIELD_SIZE+$x","y":"{\"PREV_RAND\":\"None\"}-FIELD_SIZE+$y"}},
				{"Name":"open_cell","Args":{"x":"{\"RAND\":\"None\"}-FIELD_SIZE+$x","y":"{\"PREV_RAND\":\"None\"}-FIELD_SIZE+$y"}}
			]`,
		}
		params := map[string]interface{}{"x": 2, "y": 2, "FIELD_SIZE": 10}

		// Обёртка над RunScript для логирования координат
		actions, err := ParseScript(item.Script)
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
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
			if name == "OPEN_CELL" {
				x, okX := toFloat(action.Args["x"])
				y, okY := toFloat(action.Args["y"])
				if okX && okY && int(x) >= 0 && int(x) < 10 && int(y) >= 0 && int(y) < 10 {
					res := game.OpenCell(int(x), int(y), state)
					t.Logf("Слон: открываю (%d, %d) — %s", int(x), int(y), res)
				} else {
					t.Logf("Слон: попытка открыть вне поля (%v, %v)", action.Args["x"], action.Args["y"])
				}
			}
		}
		t.Log("Слон:")
		printField(state.Field)
	}

	// Ферзь (разные типы выражений)
	{
		state := game.NewGameState()
		item := Item{
			ID: 7,
			Script: `[
				{"Name":"open_cell","Args":{"x":"$x","y":"$y"}},
				{"Name":"open_cell","Args":{"x":"{\"RAND\":\"None\"}-FIELD_SIZE+$x","y":"{\"PREV_RAND\":\"None\"}-FIELD_SIZE+$y"}},
				{"Name":"open_cell","Args":{"x":"{\"RAND\":\"None\"}-FIELD_SIZE+$x","y":"$y-{\"PREV_RAND\":\"None\"}+FIELD_SIZE"}},
				{"Name":"open_cell","Args":{"x":"{\"RAND\":\"None\"}","y":"$y"}},
				{"Name":"open_cell","Args":{"x":"$x","y":"{\"RAND\":\"None\"}"}}
			]`,
		}
		params := map[string]interface{}{"x": 3, "y": 3, "FIELD_SIZE": 10}
		_, err := UseItem(7, state, []Item{item}, params)
		if err != nil {
			t.Errorf("Ферзь: %v", err)
		}
		t.Log("Ферзь:")
		printField(state.Field)
	}
}
