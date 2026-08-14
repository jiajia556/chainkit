package exampleConfig

import (
	"github.com/spf13/cobra"
)

var exampleConfigCmd = &cobra.Command{
	Use:   "ex-config",
	Short: "Generate an example chainkit config file (chainkitExampleConfig.json)",
	Long: `Generate a minimal example ChainKit configuration file and write it to
"./chainkitExampleConfig.json".

This command does not connect to MySQL.
If the output file already exists, it will be overwritten.
`,
	Example: `  # Generate example config in the current directory
  chainkit gen ex-config
`,
	Run: func(cmd *cobra.Command, args []string) {
		genExampleConfig()
	},
}

func GetCommand() *cobra.Command {
	return exampleConfigCmd
}
