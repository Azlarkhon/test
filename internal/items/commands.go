package items

import (
	"errors"
	"lesta-battleship/server-core/internal/game"
)

// OpenCellCommand - команда для открытия ячейки
type OpenCellCommand struct {
	X, Y      int
	prevState game.CellState
}

func NewOpenCellCommand(x, y int) *OpenCellCommand {
	return &OpenCellCommand{X: x, Y: y}
}

func (cmd *OpenCellCommand) Apply(gs *game.GameState) error {
	if cmd.X < 0 || cmd.X >= 10 || cmd.Y < 0 || cmd.Y >= 10 {
		return errors.New("out of bounds")
	}
	cmd.prevState = gs.Field[cmd.X][cmd.Y]
	game.OpenCell(cmd.X, cmd.Y, gs)
	return nil
}

func (cmd *OpenCellCommand) Undo(gs *game.GameState) {
	if cmd.X >= 0 && cmd.X < 10 && cmd.Y >= 0 && cmd.Y < 10 {
		gs.Field[cmd.X][cmd.Y] = cmd.prevState
	}
}

// SetCellStatusCommand - команда для установки статуса ячейки
type SetCellStatusCommand struct {
	X, Y      int
	Status    game.CellState
	prevState game.CellState
}

func NewSetCellStatusCommand(x, y int, status game.CellState) *SetCellStatusCommand {
	return &SetCellStatusCommand{X: x, Y: y, Status: status}
}

func (cmd *SetCellStatusCommand) Apply(gs *game.GameState) error {
	if cmd.X < 0 || cmd.X >= 10 || cmd.Y < 0 || cmd.Y >= 10 {
		return errors.New("out of bounds")
	}
	cmd.prevState = gs.Field[cmd.X][cmd.Y]
	gs.Field[cmd.X][cmd.Y] = cmd.Status
	return nil
}

func (cmd *SetCellStatusCommand) Undo(gs *game.GameState) {
	if cmd.X >= 0 && cmd.X < 10 && cmd.Y >= 0 && cmd.Y < 10 {
		gs.Field[cmd.X][cmd.Y] = cmd.prevState
	}
}

// ShootCommand - команда для выстрела
type ShootCommand struct {
	Target    game.Coord
	prevState game.CellState
}

func NewShootCommand(x, y int) *ShootCommand {
	return &ShootCommand{Target: game.Coord{X: x, Y: y}}
}

func (cmd *ShootCommand) Apply(gs *game.GameState) error {
	if cmd.Target.X < 0 || cmd.Target.X >= 10 || cmd.Target.Y < 0 || cmd.Target.Y >= 10 {
		return errors.New("out of bounds")
	}
	cmd.prevState = gs.Field[cmd.Target.X][cmd.Target.Y]

	// Логика выстрела
	if gs.Field[cmd.Target.X][cmd.Target.Y] == game.ShipCell {
		gs.Field[cmd.Target.X][cmd.Target.Y] = game.Hit
	} else if gs.Field[cmd.Target.X][cmd.Target.Y] == game.Empty {
		gs.Field[cmd.Target.X][cmd.Target.Y] = game.Miss
	}
	return nil
}

func (cmd *ShootCommand) Undo(gs *game.GameState) {
	if cmd.Target.X >= 0 && cmd.Target.X < 10 && cmd.Target.Y >= 0 && cmd.Target.Y < 10 {
		gs.Field[cmd.Target.X][cmd.Target.Y] = cmd.prevState
	}
}

// SetShipCoordinatesCommand - команда для перемещения корабля
type SetShipCoordinatesCommand struct {
	OldCoords []game.Coord
	NewCoords []game.Coord
	ShipID    string
	prevField [10][10]game.CellState
}

func NewSetShipCoordinatesCommand(shipID string, oldCoords, newCoords []game.Coord) *SetShipCoordinatesCommand {
	return &SetShipCoordinatesCommand{
		ShipID:    shipID,
		OldCoords: oldCoords,
		NewCoords: newCoords,
	}
}

func (cmd *SetShipCoordinatesCommand) Apply(gs *game.GameState) error {
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
		gs.Field[coord.X][coord.Y] = game.Empty
	}

	// Устанавливаем корабль на новые позиции
	for _, coord := range cmd.NewCoords {
		gs.Field[coord.X][coord.Y] = game.ShipCell
	}

	// Обновляем координаты корабля
	if ship, exists := gs.Ships[cmd.ShipID]; exists {
		ship.Coords = cmd.NewCoords
		gs.Ships[cmd.ShipID] = ship
	}

	return nil
}

func (cmd *SetShipCoordinatesCommand) Undo(gs *game.GameState) {
	// Восстанавливаем состояние поля
	gs.Field = cmd.prevField

	// Восстанавливаем старые координаты корабля
	if ship, exists := gs.Ships[cmd.ShipID]; exists {
		ship.Coords = cmd.OldCoords
		gs.Ships[cmd.ShipID] = ship
	}
}
