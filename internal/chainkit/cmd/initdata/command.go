package initdata

import (
	"github.com/jiajia556/chainkit/internal/chainkit/config"
	"github.com/jiajia556/chainkit/pkg/utils"
	"github.com/jiajia556/tool-box/mysqlx"
	"github.com/spf13/cobra"
)

var initDataCmd = &cobra.Command{
	Use:   "init-data",
	Short: "Initialize seed data in MySQL (chains/tokens/contracts/collect/deposit)",
	Long: `Initialize ChainKit seed data into MySQL.

Requirements:
  - --config/-c (root flag) must point to a config file containing MySQL settings.
  - Mysql.AutoCreateTable must be enabled.

You must choose at least one category flag (e.g. --chains).
`,
	Example: `  # Initialize chains and tokens
	chainkit init-data --chains --tokens --config ./chainkitConfig.json

  # Initialize collect and deposit related tables
	chainkit init-data --collect-config --collect-tokens --deposit-tokens --config ./chainkitConfig.json
`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath, _ := cmd.Root().PersistentFlags().GetString("config")
		err := config.Load(configPath)
		if err != nil {
			utils.OutputFatal("Load config file failed", err)
		}
		conf := config.GetConfig()
		if !conf.Mysql.AutoCreateTable {
			utils.OutputFatal("AutoCreateTable is disabled in config, cannot initialize data")
		}
		err = mysqlx.InitMysql(conf.Mysql)
		if err != nil {
			utils.OutputFatal("Init MySQL failed", err)
		}
		initChains, _ := cmd.Flags().GetBool("chains")
		initTokens, _ := cmd.Flags().GetBool("tokens")
		initContracts, _ := cmd.Flags().GetBool("contracts")
		initCollectConfig, _ := cmd.Flags().GetBool("collect-config")
		initCollectTokens, _ := cmd.Flags().GetBool("collect-tokens")
		initDepositTokens, _ := cmd.Flags().GetBool("deposit-tokens")
		if !initChains && !initTokens && !initContracts && !initCollectConfig && !initCollectTokens && !initDepositTokens {
			utils.OutputFatal("at least one of the following flags must be set: --chains, --tokens, --contracts, --collect-config, --collect-tokens, --deposit-tokens")
		}
		initData(conf, initChains, initTokens, initContracts, initCollectConfig, initCollectTokens, initDepositTokens)
	},
}

func GetCommand() *cobra.Command {
	return initDataCmd
}

func init() {
	initDataCmd.Flags().BoolP("chains", "", false, "init chains")
	initDataCmd.Flags().BoolP("tokens", "", false, "init tokens")
	initDataCmd.Flags().BoolP("contracts", "", false, "init contracts")
	initDataCmd.Flags().BoolP("collect-config", "", false, "init collect config")
	initDataCmd.Flags().BoolP("collect-tokens", "", false, "init collect tokens")
	initDataCmd.Flags().BoolP("deposit-tokens", "", false, "init deposit tokens")
}
