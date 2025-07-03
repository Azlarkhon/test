package game

import (
	"errors"
	"fmt"
)

type PlaceShipCommand struct {
	Ship Ship
}

func (c *PlaceShipCommand) Apply(gs *GameState) error {
	definition, ok := AllowedShips[c.Ship.Type]
	if !ok {
		return fmt.Errorf("invalid ship type: %s", c.Ship.Type)
	}

	if len(c.Ship.Coords) != definition.Size {
		return fmt.Errorf("ship type %s must have size %d", c.Ship.Type, definition.Size)
	}

	countByType := make(map[ShipType]int)
	for _, s := range gs.Ships {
		countByType[s.Type]++
	}
	if countByType[c.Ship.Type] >= definition.Count {
		return fmt.Errorf("only %d %s(s) allowed", definition.Count, c.Ship.Type)
	}

	if !gs.isValidShip(c.Ship) {
		return errors.New("ship must be a straight line")
	}

	for _, coord := range c.Ship.Coords {
		if !gs.isInside(coord) {
			return errors.New("ship is out of bounds")
		}
		if !gs.isCellEmpty(coord) {
			return errors.New("cell is not empty")
		}
	}
	if gs.hasNearShips(c.Ship.Coords) {
		return errors.New("ships cannot be adjacent")
	}

	// Auto-generate ID
	gs.shipIDSeq++
	c.Ship.ID = fmt.Sprintf("%d", gs.shipIDSeq)

	// Place ship
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
