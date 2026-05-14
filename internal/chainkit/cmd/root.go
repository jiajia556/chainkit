package cmd

import (
	"github.com/jiajia556/chainkit/internal/chainkit/cmd/create"
	"github.com/jiajia556/chainkit/internal/chainkit/cmd/gen"
	"github.com/jiajia556/chainkit/internal/chainkit/cmd/initdata"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "chainkit",
	Short: "ChainKit command line tools",
	Long: `chainkit is a CLI toolbox for managing ChainKit data and keys.

Most commands that touch MySQL require a config file via --config/-c.
Use "chainkit <command> --help" for more information about a command.
`,
	Example: `  # Show help
  chainkit --help

  # Initialize seed data (requires a config file)
	chainkit init-data --chains --config ./chainkitConfig.json
`,
	Version: "1.0.0",
}

// GetRootCmd returns the root cobra command.
func GetRootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "./chainkit_config.json", "Path to configuration file")
	rootCmd.AddCommand(
		initdata.GetCommand(),
		create.GetCommand(),
		gen.GetCommand(),
	)
}
