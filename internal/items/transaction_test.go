package items

import (
	"lesta-battleship/server-core/internal/game"
	"testing"
)

func TestUseItemWithTransaction(t *testing.T) {
	// Тест 1: Простое использование предмета с транзакцией
	{
		state := game.NewGameState()
		item := Item{
			ID: 1,
			Script: `[
				{"Name":"open_cell","Args":{"x":"$x","y":"$y"}},
				{"Name":"END_PLAYER_ACTION","Args":{}}
			]`,
		}
		params := map[string]interface{}{"x": 5, "y": 5}
		itemsMap := map[int]*Item{1: &item}

		result, err := UseItem(1, state, itemsMap, params)
		if err != nil {
			t.Errorf("UseItem failed: %v", err)
		}
		if result != "item_used_successfully" {
			t.Errorf("Expected 'item_used_successfully', got '%s'", result)
		}

		// Проверяем, что ячейка была открыта
		if state.Field[5][5] != game.Revealed {
			t.Errorf("Expected cell to be revealed, got %v", state.Field[5][5])
		}
	}

	// Тест 2: Предмет с несколькими действиями
	{
		state := game.NewGameState()
		item := Item{
			ID: 2,
			Script: `[
				{"Name":"open_cell","Args":{"x":"$x","y":"$y"}},
				{"Name":"open_cell","Args":{"x":"$x+1","y":"$y"}},
				{"Name":"open_cell","Args":{"x":"$x","y":"$y+1"}},
				{"Name":"END_PLAYER_ACTION","Args":{}}
			]`,
		}
		params := map[string]interface{}{"x": 3, "y": 3}
		itemsMap := map[int]*Item{2: &item}

		result, err := UseItem(2, state, itemsMap, params)
		if err != nil {
			t.Errorf("UseItem failed: %v", err)
		}
		if result != "item_used_successfully" {
			t.Errorf("Expected 'item_used_successfully', got '%s'", result)
		}

		// Проверяем, что все ячейки были открыты
		expectedRevealed := []struct{ x, y int }{
			{3, 3}, {4, 3}, {3, 4},
		}
		for _, pos := range expectedRevealed {
			if state.Field[pos.x][pos.y] != game.Revealed {
				t.Errorf("Expected cell (%d,%d) to be revealed, got %v", pos.x, pos.y, state.Field[pos.x][pos.y])
			}
		}
	}

	// Тест 3: Предмет с SET_CELL_STATUS
	{
		state := game.NewGameState()
		item := Item{
			ID: 3,
			Script: `[
				{"Name":"SET_CELL_STATUS","Args":{"x":"$x","y":"$y","status":"ship"}},
				{"Name":"END_PLAYER_ACTION","Args":{}}
			]`,
		}
		params := map[string]interface{}{"x": 2, "y": 3}
		itemsMap := map[int]*Item{3: &item}

		result, err := UseItem(3, state, itemsMap, params)
		if err != nil {
			t.Errorf("UseItem failed: %v", err)
		}
		if result != "item_used_successfully" {
			t.Errorf("Expected 'item_used_successfully', got '%s'", result)
		}

		// Проверяем, что статус ячейки был изменен
		if state.Field[2][3] != game.ShipCell {
			t.Errorf("Expected cell to be ship, got %v", state.Field[2][3])
		}
	}

	t.Log("All transaction tests passed!")
}
