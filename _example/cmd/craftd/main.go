package main

import (
	"github.com/edsonmichaque/craft/internal/commands/craftd"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
