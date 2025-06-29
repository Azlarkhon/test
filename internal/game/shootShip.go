package game

import (
	"errors"
)

type ShootCommand struct {
	Target Coord
	Prev   CellState
}

func (c *ShootCommand) Apply(gs *GameState) error {
	if !gs.isInside(c.Target) {
		return errors.New("out of bounds")
	}
	c.Prev = gs.Field[c.Target.X][c.Target.Y]

	switch c.Prev {
	case ShipCell:
		gs.Field[c.Target.X][c.Target.Y] = Hit
	case Empty:
		gs.Field[c.Target.X][c.Target.Y] = Miss
	default:
		return errors.New("already shot here")
	}
	gs.ShotsMade = append(gs.ShotsMade, c.Target)
	return nil
}

func (c *ShootCommand) Undo(gs *GameState) {
	gs.Field[c.Target.X][c.Target.Y] = c.Prev
	gs.ShotsMade = gs.ShotsMade[:len(gs.ShotsMade)-1]
}
