package mnemonicAddress

import (
	"fmt"

	"github.com/howeyc/gopass"
	"github.com/jiajia556/chainkit/internal/chainkit/config"
	"github.com/jiajia556/chainkit/pkg/utils"
	"github.com/jiajia556/tool-box/mysqlx"
	"github.com/spf13/cobra"
)

var mneAddrCmd = &cobra.Command{
	Use:   "mnemonic-address",
	Short: "Create a derived address for an existing mnemonic (by index)",
	Long: `Derive an address from an existing mnemonic (stored in MySQL) and insert it into
the mnemonic-address table.

Requirements:
  - --config/-c (root flag) must point to a config file containing MySQL settings.
  - Mysql.AutoCreateTable must be enabled.
  - --mnemonic/-m (mnemonic ID) is required.
  - --password/-p is required to decrypt the mnemonic.

Password input:
  - If --password is set (any non-empty value), you will be prompted to enter the
    password securely (input hidden).

Output:
  - Prints: "id" and "address" for the created record.
`,
	Example: `  # Create address index 0 for mnemonic id 1 (will prompt for password)
	chainkit create mnemonic-address -m 1 -i 0 -p prompt -r "addr-0" --config ./chainkitConfig.json

  # Create address index 5 for mnemonic id 1
	chainkit create mnemonic-address -m 1 -i 5 -p prompt --config ./chainkitConfig.json
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
		if password != "" {
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
		mnemonicId, _ := cmd.Flags().GetUint64("mnemonic")
		if mnemonicId == 0 {
			utils.OutputFatal("MnemonicId is required")
		}
		index, _ := cmd.Flags().GetUint32("index")
		remark, _ := cmd.Flags().GetString("remark")
		createMnemonicAddress(mnemonicId, index, password, remark)
	},
}

func GetCommand() *cobra.Command {
	return mneAddrCmd
}

func init() {
	mneAddrCmd.Flags().StringP("password", "p", "", "Password for decrypting the mnemonic (set any value to prompt securely)")
	mneAddrCmd.Flags().StringP("remark", "r", "", "Remark for the derived address")
	mneAddrCmd.Flags().Uint32P("index", "i", 0, "The account index for the generated address (default 0)")
	mneAddrCmd.Flags().Uint64P("mnemonic", "m", 0, "Mnemonic ID")
}
