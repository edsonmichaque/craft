
package main

import (
	"context"
	"fmt"
	"os"
	"github.com/edsonmichaque/craft/internal/commands/craftctl"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	appCtx := craftctl.NewContext()

	if err := craftctl.Execute(ctx, appCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
