

package craftd


import (
	"context"
	"fmt"
	"log"
	"os"
	
"github.com/urfave/cli/v2"

)


func CmdServer(ctx context.Context, appCtx *Context) *cli.Command {
	var (
		port int
		host string
	)

	return &cli.Command{
		Name:  "server",
		Usage: "Start the server",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "port",
				Value:       8080,
				Usage:       "server port",
				Destination: &port,
			},
			&cli.StringFlag{
				Name:        "host",
				Value:       "0.0.0.0",
				Usage:       "server host",
				Destination: &host,
			},
		},
		Action: func(c *cli.Context) error {
			if appCtx.Debug {
				log.Println("Debug mode enabled")
			}
			return runServer(ctx, host, port)
		},
	}
}

