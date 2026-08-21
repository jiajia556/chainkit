package deposit

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

func TestSameWebSocketTokenConfig(t *testing.T) {
	address := common.HexToAddress("0x0000000000000000000000000000000000000001")
	left := &websocketTokenConfig{
		minDepositAmounts: map[common.Address]decimal.Decimal{
			address: decimal.NewFromInt(100),
		},
	}
	right := &websocketTokenConfig{
		minDepositAmounts: map[common.Address]decimal.Decimal{
			address: decimal.NewFromInt(100),
		},
	}
	if !sameWebSocketTokenConfig(left, right) {
		t.Fatal("equivalent websocket token configurations should match")
	}

	right.minDepositAmounts[address] = decimal.NewFromInt(101)
	if sameWebSocketTokenConfig(left, right) {
		t.Fatal("different minimum deposit amounts should trigger resubscription")
	}
}
