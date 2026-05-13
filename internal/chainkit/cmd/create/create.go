package create

import (
	depaddr "github.com/jiajia556/chainkit/internal/chainkit/cmd/create/depaddr"
	"github.com/jiajia556/chainkit/internal/chainkit/cmd/create/mnemonic"
	"github.com/jiajia556/chainkit/internal/chainkit/cmd/create/mnemonicAddress"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create keys and addresses",
	Long: `Create and store keys/addresses used by ChainKit.

This command groups subcommands for creating mnemonics, deriving mnemonic addresses,
and generating user deposit addresses.
`,
	Example: `  # Create a new mnemonic (will prompt for password)
	chainkit create mnemonic -p prompt --out std --remark "dev-seed" --config ./chainkitConfig.json

  # Derive an address for an existing mnemonic
	chainkit create mnemonic-address -m 1 -i 0 -p prompt --config ./chainkitConfig.json

  # Batch create deposit addresses
	chainkit create deposit-address -p prompt -n 10 --config ./chainkitConfig.json
`,
}

func GetCommand() *cobra.Command {
	return createCmd
}

func init() {
	createCmd.AddCommand(
		mnemonic.GetCommand(),
		mnemonicAddress.GetCommand(),
		depaddr.GetCommand(),
	)
}
