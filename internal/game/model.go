package game

type CellState int

const (
	Empty CellState = iota
	ShipCell
	Miss
	Hit
	Revealed
)

type Coord struct {
	X int
	Y int
}

type Ship struct {
	ID     string   `json:"id"`
	Type   ShipType `json:"type"`
	Coords []Coord  `json:"coords"`
}

type ShipType string

const (
	Battleship ShipType = "battleship" // size 4
	Cruiser    ShipType = "cruiser"    // size 3
	Destroyer  ShipType = "destroyer"  // size 2
	Submarine  ShipType = "submarine"  // size 1
)

var AllowedShips = map[ShipType]struct {
	Size  int
	Count int
}{
	Battleship: {Size: 4, Count: 1},
	Cruiser:    {Size: 3, Count: 2},
	Destroyer:  {Size: 2, Count: 3},
	Submarine:  {Size: 1, Count: 4},
}

type GameState struct {
	Field     [10][10]CellState
	Ships     map[string]Ship
	ShotsMade []Coord
	shipIDSeq int // <-- Add this line
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

func (gs *GameState) isValidShip(ship Ship) bool {
	firstX := ship.Coords[0].X
	firstY := ship.Coords[0].Y

	isHorizontal := true
	isVertical := true

	for _, coord := range ship.Coords {
		if coord.X != firstX {
			isVertical = false
		}
	}

	for _, coord := range ship.Coords {
		if coord.Y != firstY {
			isHorizontal = false
		}
	}

	return isHorizontal || isVertical
}

func (gs *GameState) hasNearShips(coords []Coord) bool {
	nearCoords := getNearCoords(coords)
	for _, coord := range nearCoords {
		if !gs.isInside(coord) {
			continue
		}
		if gs.Field[coord.X][coord.Y] == ShipCell {
			return true
		}
	}
	return false
}

func getNearCoords(coords []Coord) []Coord {
	nearCoords := []Coord{}
	for _, coord := range coords {
		nearCoords = append(nearCoords, Coord{X: coord.X - 1, Y: coord.Y})
		nearCoords = append(nearCoords, Coord{X: coord.X + 1, Y: coord.Y})
		nearCoords = append(nearCoords, Coord{X: coord.X, Y: coord.Y - 1})
		nearCoords = append(nearCoords, Coord{X: coord.X, Y: coord.Y + 1})
	}
	return nearCoords
}

func OpenCell(x, y int, gs *GameState) string {
	if x < 0 || y < 0 || x >= len(gs.Field) || y >= len(gs.Field[0]) {
		return "invalid"
	}
	switch gs.Field[x][y] {
	case Empty:
		gs.Field[x][y] = Revealed
		return "empty"
	case ShipCell:
		return "ship"
	case Miss:
		return "miss"
	case Hit:
		return "hit"
	case Revealed:
		return "revealed"
	default:
		return "unknown"
	}
}
