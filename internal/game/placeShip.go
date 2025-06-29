package game

import (
	"errors"
)

type PlaceShipCommand struct {
	Ship Ship
}

func (c *PlaceShipCommand) Apply(gs *GameState) error {
	if !gs.isValidShip(c.Ship) {
		return errors.New("invalid ship: must be a straight line")
	}

	for _, coord := range c.Ship.Coords {
		if !gs.isInside(coord) {
			return errors.New("ship is outside the field")
		}
		if !gs.isCellEmpty(coord) {
			return errors.New("ship cant be placed, cell is not empty")
		}
	}

	if gs.hasNearShips(c.Ship.Coords) {
		return errors.New("ships cannot be near")
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
