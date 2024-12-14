

package craftctl


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


// Common server implementation
func runServer(ctx context.Context, host string, port int) error {
	// Server implementation
	return nil
}


func CmdRoot(ctx context.Context, appCtx *AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "craftctl",
		Short: "craft CLI",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Setup logging, tracing, etc.
		},
	}

	cmd.PersistentFlags().StringVar(&appCtx.ConfigPath, "config", "", "config file path")
	cmd.PersistentFlags().BoolVar(&appCtx.Debug, "debug", false, "enable debug mode")

	cmd.AddCommand(
		CmdVersion(ctx, appCtx),
		CmdServer(ctx, appCtx),
	)

	return cmd
}

func Execute(ctx context.Context, appCtx *AppContext) error {
	cmd := CmdRoot(ctx, appCtx)
	return cmd.Execute()
}


