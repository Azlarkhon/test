package sample

import (
	"fmt"
	"log"

	"lesta-battleship/server-core/internal/game"
	"lesta-battleship/server-core/internal/transaction"
)

func RunSample() {
	gs := game.NewGameState()

	// Исходный корабль
	// Можешь чекнуть транзакцию сделая x или y > 9 (0-9)
	ship := game.Ship{
		ID: "ship-1",
		Coords: []game.Coord{
			{X: 1, Y: 1},
			{X: 1, Y: 2},
			{X: 1, Y: 3},
		},
	}

	// Транзакция 1: разместить корабль
	tx1 := transaction.NewTransaction()
	tx1.Add(&game.PlaceShipCommand{Ship: ship})

	if err := tx1.Execute(gs); err != nil {
		log.Fatal("place ship failed:", err)
	}

	fmt.Println("Корабль размещен")

	// Транзакция 2: переместить корабль и выстрелить

	// Новый корабль с тем же ID, но в другом месте
	// Так же и тут можно чекать
	moved := game.Ship{
		ID: "ship-1",
		Coords: []game.Coord{
			{X: 2, Y: 2},
			{X: 2, Y: 3},
			{X: 2, Y: 11},
		},
	}

	tx2 := transaction.NewTransaction()
	tx2.Add(&game.RemoveShipCommand{ShipID: "ship-1"})
	tx2.Add(&game.PlaceShipCommand{Ship: moved})
	tx2.Add(&game.ShootCommand{Target: game.Coord{X: 2, Y: 3}})

	if err := tx2.Execute(gs); err != nil {
		fmt.Println("Ошибка, все откатилось", err)
	} else {
		fmt.Println("Корабль перемещен и произведен выстрел")
	}

	// Транзакция 3: попытка разместить корабль рядом
	tx3 := transaction.NewTransaction()
	tx3.Add(&game.PlaceShipCommand{Ship: game.Ship{
		ID: "ship-2",
		Coords: []game.Coord{
			{X: 2, Y: 2},
		},
	}})

	if err := tx3.Execute(gs); err != nil {
		fmt.Println("Ошибка, все откатилось", err)
	} else {
		fmt.Println("Корабль размещен")
	}

	printBoard(gs)
}

func printBoard(gs *game.GameState) {
	fmt.Println("   0 1 2 3 4 5 6 7 8 9")
	for y := 0; y < 10; y++ {
		fmt.Printf("%d  ", y)
		for x := 0; x < 10; x++ {
			cell := gs.Field[x][y]
			var ch string
			switch cell {
			case game.Empty:
				ch = "."
			case game.ShipCell:
				ch = "S"
			case game.Miss:
				ch = "o"
			case game.Hit:
				ch = "X"
			default:
				ch = "?"
			}
			fmt.Print(ch + " ")
		}
		fmt.Println()
	}
}
