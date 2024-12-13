package main

import (
	"github.com/edsonmichaque/craft/internal/commands/craftctl"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
