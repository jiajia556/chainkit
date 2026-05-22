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
	Short: "",
	Long:  ``,
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
