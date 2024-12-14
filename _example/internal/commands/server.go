

package commands


import (
	"context"
	"fmt"
	"log"
	"os"
	
"github.com/spf13/cobra"

)


type AppContext struct {
	ConfigPath string
	Debug      bool
}

func NewAppContext() *AppContext {
	return &AppContext{}
}



func CmdServer(ctx context.Context, appCtx *AppContext) *cobra.Command {
	var (
		port int
		host string
	)

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if appCtx.Debug {
				log.Println("Debug mode enabled")
			}
			return runServer(ctx, host, port)
		},
	}

	cmd.Flags().IntVar(&port, "port", 8080, "server port")
	cmd.Flags().StringVar(&host, "host", "0.0.0.0", "server host")

	return cmd
}

