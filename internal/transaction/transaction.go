package transaction

import (
	"fmt"
	"lesta-battleship/server-core/internal/game"
)

type Command interface {
	Apply(gs *game.GameState) error
	Undo(gs *game.GameState)
}

type Transaction struct {
	commands []Command
}

func NewTransaction() *Transaction {
	return &Transaction{}
}

func (tx *Transaction) Add(cmd Command) {
	tx.commands = append(tx.commands, cmd)
}

func (tx *Transaction) Execute(gs *game.GameState) error {
	for i, cmd := range tx.commands {
		if err := cmd.Apply(gs); err != nil {
			for j := i - 1; j >= 0; j-- {
				tx.commands[j].Undo(gs)
			}
			return fmt.Errorf("error at step %d: %w", i, err)
		}
	}
	return nil
}
