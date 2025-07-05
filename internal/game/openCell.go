package game

import "errors"

// OpenCellCommand - команда для открытия ячейки
type OpenCellCommand struct {
	X, Y      int
	prevState CellState
}

func NewOpenCellCommand(x, y int) *OpenCellCommand {
	return &OpenCellCommand{X: x, Y: y}
}

func (cmd *OpenCellCommand) Apply(gs *GameState) error {
	if cmd.X < 0 || cmd.X >= 10 || cmd.Y < 0 || cmd.Y >= 10 {
		return errors.New("out of bounds")
	}
	cmd.prevState = gs.Field[cmd.X][cmd.Y]
	OpenCell(cmd.X, cmd.Y, gs)
	return nil
}

func (cmd *OpenCellCommand) Undo(gs *GameState) {
	if cmd.X >= 0 && cmd.X < 10 && cmd.Y >= 0 && cmd.Y < 10 {
		gs.Field[cmd.X][cmd.Y] = cmd.prevState
	}
}

// SetCellStatusCommand - команда для установки статуса ячейки
type SetCellStatusCommand struct {
	X, Y      int
	Status    CellState
	prevState CellState
}

func NewSetCellStatusCommand(x, y int, status CellState) *SetCellStatusCommand {
	return &SetCellStatusCommand{X: x, Y: y, Status: status}
}

func (cmd *SetCellStatusCommand) Apply(gs *GameState) error {
	if cmd.X < 0 || cmd.X >= 10 || cmd.Y < 0 || cmd.Y >= 10 {
		return errors.New("out of bounds")
	}
	cmd.prevState = gs.Field[cmd.X][cmd.Y]
	gs.Field[cmd.X][cmd.Y] = cmd.Status
	return nil
}

func (cmd *SetCellStatusCommand) Undo(gs *GameState) {
	if cmd.X >= 0 && cmd.X < 10 && cmd.Y >= 0 && cmd.Y < 10 {
		gs.Field[cmd.X][cmd.Y] = cmd.prevState
	}
}
