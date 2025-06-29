package game

import (
	"errors"
)

type PlaceShipCommand struct {
	Ship Ship
}

func (c *PlaceShipCommand) Apply(gs *GameState) error {
	for _, coord := range c.Ship.Coords {
		if !gs.isInside(coord) || !gs.isCellEmpty(coord) {
			return errors.New("invalid ship placement")
		}
	}
	for _, coord := range c.Ship.Coords {
		gs.Field[coord.X][coord.Y] = ShipCell
	}
	gs.Ships[c.Ship.ID] = c.Ship
	return nil
}

func (c *PlaceShipCommand) Undo(gs *GameState) {
	for _, coord := range c.Ship.Coords {
		gs.Field[coord.X][coord.Y] = Empty
	}
	delete(gs.Ships, c.Ship.ID)
}
