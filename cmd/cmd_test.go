package cmd

import (
	"flag"
	"fmt"
	"testing"

	"github.com/jiajia556/chainkit/internal/collect/config"
	"github.com/jiajia556/chainkit/service"
	"github.com/jiajia556/tool-box/mysqlx"
)

func Test(t *testing.T) {
	var configPath string
	flag.StringVar(&configPath, "config", "E:\\work\\gowork\\chainkit\\chainkit_config.json", "Config json file path")
	flag.Parse()
	err := config.Load(configPath)
	if err != nil {
		panic(err)
	}

	err = mysqlx.InitMysql(config.GetConfig().Mysql)
	if err != nil {
		panic(err)
	}

	password := "123123"

	srv, err := service.NewChainService(2)
	if err != nil {
		panic(err)
	}
	err = srv.SetFromByMnemonicAddress(1, password)
	if err != nil {
		panic(err)
	}

	status, err := srv.GetTxStatus("0xb79059cd85e3c6f2534538dd9a17e05a6bd21789e731b5e825890a4ad641c1d5")
	if err != nil {
		panic(err)
	}
	fmt.Println(status)

	//self, err := srv.GetFromAddress()
	//if err != nil {
	//	panic(err)
	//}
	//client := srv.GetClient()
	//nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(self))
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("nonce:", nonce)
	//occupied, err := srv.IsNonceOccupied(self, 1)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(occupied)
	//hash, nonce, err, err2 := srv.TransferETH(self, decimal.New(1, 1), service.Nonce(2))
	//fmt.Println("hash", hash, "nonce", nonce, "err", err, "err2", err2)
}
