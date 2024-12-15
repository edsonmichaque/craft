

package craftd


import (
	"context"
	"fmt"
	"log"
	"os"
	
"github.com/urfave/cli/v2"
"github.com/edsonmichaque/craft/pkg/version"

)


func CmdVersion(ctx context.Context, appCtx *Context) *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Print version information",
		Action: func(c *cli.Context) error {
			fmt.Printf("Version: %s\n", version.Version)
			fmt.Printf("Commit: %s\n", version.Commit)
			fmt.Printf("Build Date: %s\n", version.BuildDate)
			return nil
		},
	}
}

