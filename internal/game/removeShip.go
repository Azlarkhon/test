package game

import "errors"

type RemoveShipCommand struct {
	ShipID string
	Backup Ship
}

func (c *RemoveShipCommand) Apply(gs *GameState) error {
	ship, ok := gs.Ships[c.ShipID]
	if !ok {
		return errors.New("ship not found")
	}
	c.Backup = ship
	for _, coord := range ship.Coords {
		gs.Field[coord.X][coord.Y] = Empty
	}
	delete(gs.Ships, c.ShipID)
	return nil
}

func (c *RemoveShipCommand) Undo(gs *GameState) {
	for _, coord := range c.Backup.Coords {
		gs.Field[coord.X][coord.Y] = ShipCell
	}
	gs.Ships[c.Backup.ID] = c.Backup
}
