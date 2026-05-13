package gen

import (
	"github.com/jiajia556/chainkit/internal/chainkit/cmd/gen/exampleConfig"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate helper files",
	Long: `Generate helper files for ChainKit, such as example configuration files.
`,
	Example: `  # Generate an example config file in the current directory
  chainkit gen ex-config
`,
}

func GetCommand() *cobra.Command {
	return genCmd
}

func init() {
	genCmd.AddCommand(
		exampleConfig.GetCommand(),
	)
}
