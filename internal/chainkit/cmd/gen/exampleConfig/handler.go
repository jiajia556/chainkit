package exampleConfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jiajia556/chainkit/internal/chainkit/config"
	"github.com/jiajia556/chainkit/models/chainkitchains"
	"github.com/jiajia556/chainkit/models/chainkitcollectconfig"
	"github.com/jiajia556/chainkit/models/chainkitcollecttokens"
	"github.com/jiajia556/chainkit/models/chainkitcontracts"
	"github.com/jiajia556/chainkit/models/chainkitdeposittokens"
	"github.com/jiajia556/chainkit/models/chainkittokens"
	"github.com/jiajia556/tool-box/mysqlx"
	"github.com/shopspring/decimal"
)

func genExampleConfig() {
	now := time.Now().UTC()
	conf := config.Config{
		Mysql: mysqlx.MysqlConfig{},
		MysqlInit: config.MysqlInit{
			Chains: []*chainkitchains.ChainChains{
				{
					Name:              "Ethereum",
					Rpc:               "https://rpc.ankr.com/eth",
					ChainId:           1,
					SafeConfirmations: 12,
				},
			},
			Tokens: []*chainkittokens.ChainTokens{
				{
					ChainDbId:       1,
					ContractAddress: "0x0000000000000000000000000000000000000000",
					Logo:            "",
					Symbol:          "ETH",
					Decimals:        18,
					Remark:          "native token",
					CreatedAt:       now,
					UpdatedAt:       now,
				},
			},
			Contracts: []*chainkitcontracts.ChainContracts{
				{
					ChainDbId: 1,
					Name:      "MultiTransfer",
					Address:   "0x0000000000000000000000000000000000000000",
					Remark:    "example contract",
				},
			},
			CollectConfig: []*chainkitcollectconfig.ChainCollectConfig{
				{
					Id:                           1,
					ChainDbId:                    1,
					GasProviderMnemonicAddressId: 1,
					DefaultCollectToAddress:      "0x0000000000000000000000000000000000000000",
					DefaultErc20TransferGasLimit: decimal.NewFromInt(70000),
				},
			},
			CollectTokens: []*chainkitcollecttokens.ChainCollectTokens{
				{
					ChainDbId:        1,
					TokenId:          1,
					TokenAddress:     "0x0000000000000000000000000000000000000000",
					Symbol:           "USDT",
					Decimals:         6,
					MinCollectAmount: decimal.NewFromInt(1000000),
					TransferGasLimit: decimal.NewFromInt(70000),
					ToAddress:        "",
					Status:           1,
					Remark:           "example collect token",
					CreatedAt:        now,
					UpdatedAt:        now,
				},
			},
			DepositTokens: []*chainkitdeposittokens.ChainDepositTokens{
				{
					ChainDbId:      1,
					TokenId:        1,
					Status:         1,
					InitStartBlock: 0,
				},
			},
		},
	}

	b, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		panic(err)
	}
	outPath := filepath.Join(".", "chainkitExampleConfig.json")
	if err := os.WriteFile(outPath, b, 0o644); err != nil {
		panic(err)
	}
	fmt.Println("generated:", outPath)
}
