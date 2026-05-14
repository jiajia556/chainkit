package depaddr

import (
	"fmt"

	"github.com/howeyc/gopass"
	"github.com/jiajia556/chainkit/internal/chainkit/config"
	"github.com/jiajia556/chainkit/pkg/utils"
	"github.com/jiajia556/tool-box/mysqlx"
	"github.com/spf13/cobra"
)

var depAddrCmd = &cobra.Command{
	Use:   "deposit-address",
	Short: "Batch create user deposit addresses (encrypted private keys stored in MySQL)",
	Long: `Create one or more deposit addresses and store them into MySQL.

For each created address, the private key is encrypted with the given password and
stored in table: chain_user_deposit_address.

Requirements:
  - --config (root flag) must point to a config file containing MySQL settings.
  - Mysql.AutoCreateTable must be enabled.
  - --password/-p is required.

Password input:
  - If --password is set (any non-empty value), you will be prompted to enter the
    password securely (input hidden).

Notes:
  - This command does not print generated addresses by default.
`,
	Example: `  # Create 1 deposit address (will prompt for password)
	chainkit create deposit-address --password prompt --count 1 --remark "dep" --config ./chainkitConfig.json

  # Create 10 deposit addresses
	chainkit create deposit-address --password prompt --count 10 --config ./chainkitConfig.json
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
		password, _ := cmd.Flags().GetString("password")
		if password == "" {
			fmt.Println("enter password:")
			passwordByte, err := gopass.GetPasswd()
			if err != nil {
				panic(err)
			}
			password = string(passwordByte)
		}
		if password == "" {
			utils.OutputFatal("Password is required")
		}
		remark, _ := cmd.Flags().GetString("remark")
		count, _ := cmd.Flags().GetInt("count")
		createDepositAddress(count, password, remark)
	},
}

func GetCommand() *cobra.Command {
	return depAddrCmd
}

func init() {
	depAddrCmd.Flags().StringP("password", "p", "", "Encryption password (set any value to prompt securely)")
	depAddrCmd.Flags().StringP("remark", "r", "", "Remark for generated addresses")
	depAddrCmd.Flags().IntP("count", "n", 1, "The number of deposit addresses to create")
}
