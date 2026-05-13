package initdata

import (
	"fmt"

	"github.com/jiajia556/chainkit/internal/chainkit/config"
	"github.com/jiajia556/chainkit/models/chainkitchains"
	"github.com/jiajia556/chainkit/models/chainkitcollectconfig"
	"github.com/jiajia556/chainkit/models/chainkitcontracts"
	"github.com/jiajia556/chainkit/models/chainkitdeposittokens"
	"github.com/jiajia556/chainkit/models/chainkittokens"
)

func initData(conf *config.Config, initChains, initTokens, initContracts, initCollectConfig, initCollectTokens, initDepositTokens bool) {
	if initChains {
		for _, chainConf := range conf.MysqlInit.Chains {
			chain := chainkitchains.NewRecord()
			chain.SetModel(chainConf)
			err := chain.Create()
			if err != nil {
				fmt.Println("failed to create chain:", chain.Model, "error: ", err)
			}
		}
	}

	if initTokens {
		for _, tokenConf := range conf.MysqlInit.Tokens {
			token := chainkittokens.NewRecord()
			token.SetModel(tokenConf)
			err := token.Create()
			if err != nil {
				fmt.Println("failed to create token:", token.Model, "error: ", err)
			}
		}
	}

	if initContracts {
		for _, contractConf := range conf.MysqlInit.Contracts {
			contract := chainkitcontracts.NewRecord()
			contract.SetModel(contractConf)
			err := contract.Create()
			if err != nil {
				fmt.Println("failed to create contract:", contract.Model, "error: ", err)
			}
		}
	}

	if initCollectConfig {
		for _, collectConf := range conf.MysqlInit.CollectConfig {
			collectConfig := chainkitcollectconfig.NewRecord()
			collectConfig.SetModel(collectConf)
			err := collectConfig.Create()
			if err != nil {
				fmt.Println("failed to create collect config:", collectConfig.Model, "error: ", err)
			}
		}
	}

	if initCollectTokens {
		for _, collectTokenConf := range conf.MysqlInit.CollectTokens {
			collectToken := chainkittokens.NewRecord()
			collectToken.SetModel(collectTokenConf)
			err := collectToken.Create()
			if err != nil {
				fmt.Println("failed to create collect token:", collectToken.Model, "error: ", err)
			}
		}
	}

	if initDepositTokens {
		for _, depositTokenConf := range conf.MysqlInit.DepositTokens {
			depositToken := chainkitdeposittokens.NewRecord()
			depositToken.SetModel(depositTokenConf)
			err := depositToken.Create()
			if err != nil {
				fmt.Println("failed to create collect token:", depositToken.Model, "error: ", err)
			}
		}
	}
}
