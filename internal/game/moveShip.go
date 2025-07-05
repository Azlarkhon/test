package game

import "errors"

// SetShipCoordinatesCommand - команда для перемещения корабля
type SetShipCoordinatesCommand struct {
	OldCoords []Coord
	NewCoords []Coord
	ShipID    string
	prevField [10][10]CellState
}

func NewSetShipCoordinatesCommand(shipID string, oldCoords, newCoords []Coord) *SetShipCoordinatesCommand {
	return &SetShipCoordinatesCommand{
		ShipID:    shipID,
		OldCoords: oldCoords,
		NewCoords: newCoords,
	}
}

func (cmd *SetShipCoordinatesCommand) Apply(gs *GameState) error {
	// Сохраняем текущее состояние поля
	cmd.prevField = gs.Field

	// Проверяем границы для новых координат
	for _, coord := range cmd.NewCoords {
		if coord.X < 0 || coord.X >= 10 || coord.Y < 0 || coord.Y >= 10 {
			return errors.New("out of bounds")
		}
	}

	// Очищаем старые позиции корабля
	for _, coord := range cmd.OldCoords {
		gs.Field[coord.X][coord.Y] = Empty
	}

	// Устанавливаем корабль на новые позиции
	for _, coord := range cmd.NewCoords {
		gs.Field[coord.X][coord.Y] = ShipCell
	}

	// Обновляем координаты корабля
	if ship, exists := gs.Ships[cmd.ShipID]; exists {
		ship.Coords = cmd.NewCoords
		gs.Ships[cmd.ShipID] = ship
	}

	return nil
}

func (cmd *SetShipCoordinatesCommand) Undo(gs *GameState) {
	// Восстанавливаем состояние поля
	gs.Field = cmd.prevField

	// Восстанавливаем старые координаты корабля
	if ship, exists := gs.Ships[cmd.ShipID]; exists {
		ship.Coords = cmd.OldCoords
		gs.Ships[cmd.ShipID] = ship
	}
}
