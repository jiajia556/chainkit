package mnemonic

import (
	"fmt"

	"github.com/howeyc/gopass"
	"github.com/jiajia556/chainkit/internal/chainkit/config"
	"github.com/jiajia556/chainkit/pkg/utils"
	"github.com/jiajia556/tool-box/mysqlx"
	"github.com/spf13/cobra"
)

var mneCmd = &cobra.Command{
	Use:   "mnemonic",
	Short: "Create a new mnemonic phrase (stored encrypted in MySQL)",
	Long: `Create a new BIP-39 mnemonic phrase and store it into MySQL.

The mnemonic words are encrypted before being stored (table: chain_mnemonics).

Password:
	  - --password/-p is required to encrypt the mnemonic.
  - If --password is set (any non-empty value), you will be prompted to enter the
	password securely (input hidden).

Output:
  - By default (--out is empty), this command only stores the mnemonic in MySQL
	and does NOT print the words.
  - Use --out std to print "id" and "words" to stdout.
  - Use --out <file> to write the output to a file.
`,
	Example: `  # Store mnemonic in MySQL only (do not print words)
	chainkit create mnemonic -p prompt --remark "prod-seed" --config ./chainkitConfig.json

  # Print mnemonic to stdout
	chainkit create mnemonic -p prompt --out std --remark "dev-seed" --config ./chainkitConfig.json

  # Prompt for password and print mnemonic to stdout
	chainkit create mnemonic --password prompt --out std --remark "encrypted" --config ./chainkitConfig.json

  # Save mnemonic to a file
	chainkit create mnemonic -p prompt --out ./mnemonic.txt --remark "backup" --config ./chainkitConfig.json
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
		out, _ := cmd.Flags().GetString("out")
		remark, _ := cmd.Flags().GetString("remark")
		createMnemonic(password, out, remark)
	},
}

func GetCommand() *cobra.Command {
	return mneCmd
}

func init() {
	mneCmd.Flags().StringP("password", "p", "", "Encryption password (set any value to prompt securely)")
	mneCmd.Flags().StringP("out", "o", "", "The path to the output mnemonic phrase file or std")
	mneCmd.Flags().StringP("remark", "r", "", "Remark for the mnemonic")
}
