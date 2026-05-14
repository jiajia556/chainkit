package config

import (
	"github.com/jiajia556/chainkit/internal/common/config"
	"github.com/jiajia556/chainkit/models/chainkitchains"
	"github.com/jiajia556/chainkit/models/chainkitcollectconfig"
	"github.com/jiajia556/chainkit/models/chainkitcollecttokens"
	"github.com/jiajia556/chainkit/models/chainkitcontracts"
	"github.com/jiajia556/chainkit/models/chainkitdeposittokens"
	"github.com/jiajia556/chainkit/models/chainkittokengroups"
	"github.com/jiajia556/chainkit/models/chainkittokens"

	"github.com/jiajia556/tool-box/mysqlx"
)

type Config struct {
	Mysql     mysqlx.MysqlConfig `json:"mysql" yaml:"mysql"`
	MysqlInit MysqlInit          `json:"mysql_init" yaml:"mysql_init"`
}

type MysqlInit struct {
	Chains        []*chainkitchains.ChainChains               `json:"chains" yaml:"chains"`
	TokenGroups   []*chainkittokengroups.ChainTokenGroups     `json:"token_groups" yaml:"token_groups"`
	Tokens        []*chainkittokens.ChainTokens               `json:"tokens" yaml:"tokens"`
	Contracts     []*chainkitcontracts.ChainContracts         `json:"contracts" yaml:"contracts"`
	CollectConfig []*chainkitcollectconfig.ChainCollectConfig `json:"collect_config" yaml:"collect_config"`
	CollectTokens []*chainkitcollecttokens.ChainCollectTokens `json:"collect_tokens" yaml:"collect_tokens"`
	DepositTokens []*chainkitdeposittokens.ChainDepositTokens `json:"deposit_tokens" yaml:"deposit_tokens"`
}

var cfg *config.ConfigManager[Config]

func Load(path string) error {
	cfg = config.NewManager[Config]()
	return cfg.Load(path)
}

func GetConfig() *Config {
	res := cfg.Get()
	return &res
}

func CreateConfigFile(path string) error {
	cfg = config.NewManager[Config]()
	return cfg.Save(path)
}
