

package craftctl


import (
	"context"
	"fmt"
	"log"
	"os"
	
"github.com/urfave/cli/v2"

)


type Context struct {
	ConfigPath string
	Debug      bool
}

func NewContext() *Context {
	return &Context{}
}

func CmdRoot(ctx context.Context, appCtx *Context) *cli.App {
	return &cli.App{
		Name:  "craftctl",
		Usage: "craft CLI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Usage:       "config file path",
				Destination: &appCtx.ConfigPath,
			},
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "enable debug mode",
				Destination: &appCtx.Debug,
			},
		},
		Before: func(c *cli.Context) error {
			// Setup logging, tracing, etc.
			return nil
		},
		Commands: []*cli.Command{
			CmdVersion(ctx, appCtx),
			CmdServer(ctx, appCtx),
		},
	}
}

func Execute(ctx context.Context, appCtx *Context) error {
	return CmdRoot(ctx, appCtx).Run(os.Args)
}

