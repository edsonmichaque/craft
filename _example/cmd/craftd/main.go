
package main

import (
	"context"
	"fmt"
	"os"
	"github.com/edsonmichaque/craft/internal/commands/craftd"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	appCtx := craftd.NewContext()

	if err := craftd.Execute(ctx, appCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
