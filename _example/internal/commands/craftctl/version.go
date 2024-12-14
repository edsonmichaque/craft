

package craftctl


import (
	"context"
	"fmt"
	"log"
	"os"
	
"github.com/spf13/cobra"
"github.com/edsonmichaque/craft/pkg/version"

)


type AppContext struct {
	ConfigPath string
	Debug      bool
}

func NewAppContext() *AppContext {
	return &AppContext{}
}



func CmdVersion(ctx context.Context, appCtx *AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Version: %s\n", version.Version)
			fmt.Printf("Commit: %s\n", version.Commit)
			fmt.Printf("Build Date: %s\n", version.BuildDate)
			return nil
		},
	}
}


