package eip7702batch

import (
	"testing"

	"github.com/jiajia556/chainkit/models/chainkitcollectconfig"
)

func TestValidateAutomaticConfig(t *testing.T) {
	valid := chainkitcollectconfig.ChainCollectConfig{
		ChainDbId:                    1,
		GasProviderMnemonicAddressId: 10,
		EIP7702DelegateAddress:       "0x1111111111111111111111111111111111111111",
		EIP7702ExecutorAddress:       "0x2222222222222222222222222222222222222222",
		EIP7702MaxBatchItems:         100,
	}
	if err := validateAutomaticConfig(&valid); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		mutate func(*chainkitcollectconfig.ChainCollectConfig)
	}{
		{name: "missing sponsor", mutate: func(c *chainkitcollectconfig.ChainCollectConfig) { c.GasProviderMnemonicAddressId = 0 }},
		{name: "invalid delegate", mutate: func(c *chainkitcollectconfig.ChainCollectConfig) { c.EIP7702DelegateAddress = "bad" }},
		{name: "invalid executor", mutate: func(c *chainkitcollectconfig.ChainCollectConfig) { c.EIP7702ExecutorAddress = "bad" }},
		{name: "oversized batch", mutate: func(c *chainkitcollectconfig.ChainCollectConfig) { c.EIP7702MaxBatchItems = 101 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := valid
			test.mutate(&config)
			if err := validateAutomaticConfig(&config); err == nil {
				t.Fatal("expected invalid config error")
			}
		})
	}
}
