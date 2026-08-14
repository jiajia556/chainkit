package _import

import (
	"github.com/jiajia556/chainkit/internal/chainkit/cmd/import/mnemonic"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "import",
	Short: "",
	Long:  ``,
}

func GetCommand() *cobra.Command {
	return createCmd
}

func init() {
	createCmd.AddCommand(
		mnemonic.GetCommand(),
	)
}
