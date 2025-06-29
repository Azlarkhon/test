package game

type CellState int

const (
	Empty CellState = iota
	ShipCell
	Miss
	Hit
)

type Coord struct {
	X int
	Y int
}

type Ship struct {
	ID     string
	Coords []Coord
}

type GameState struct {
	Field     [10][10]CellState
	Ships     map[string]Ship
	ShotsMade []Coord
}

func NewGameState() *GameState {
	return &GameState{
		Ships: make(map[string]Ship),
	}
}

func (gs *GameState) isInside(c Coord) bool {
	return c.X >= 0 && c.X < 10 && c.Y >= 0 && c.Y < 10
}

func (gs *GameState) isCellEmpty(c Coord) bool {
	return gs.Field[c.X][c.Y] == Empty
}
