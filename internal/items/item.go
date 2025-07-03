package items

import (
	"encoding/json"
	"fmt"
	"lesta-battleship/server-core/internal/game"
	"net/http"
)

type Item struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
	Script      string `json:"script"`
	ID          int    `json:"id"`
}

func GetItemsInfo() ([]Item, error) {
	r, err := http.Get("http://37.9.53.107/items/")
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var items []Item
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		return nil, err
	}
	return items, nil
}

func UseItem(id int, state *game.GameState, itemsList []Item, params map[string]interface{}) (string, error) {
	var item *Item
	for i := range itemsList {
		if itemsList[i].ID == id {
			item = &itemsList[i]
			break
		}
	}
	if item == nil {
		return "", fmt.Errorf("item with id %d not found", id)
	}
	return RunScript(item.Script, state, params)
}
