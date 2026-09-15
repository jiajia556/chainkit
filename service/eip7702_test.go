package service

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jiajia556/chainkit/pkg/contracts/batchsweepexecutor"
)

func TestInspectDelegationCode(t *testing.T) {
	expected := common.HexToAddress("0x1111111111111111111111111111111111111111")
	other := common.HexToAddress("0x2222222222222222222222222222222222222222")

	tests := []struct {
		name     string
		code     []byte
		want     DelegationState
		delegate common.Address
	}{
		{name: "empty", want: DelegationNone},
		{name: "expected", code: types.AddressToDelegation(expected), want: DelegationExpected, delegate: expected},
		{name: "other", code: types.AddressToDelegation(other), want: DelegationOther, delegate: other},
		{name: "unsupported", code: []byte{0x60, 0x00}, want: DelegationUnsupportedCode},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := InspectDelegationCode(test.code, expected)
			if got.State != test.want || got.Delegate != test.delegate {
				t.Fatalf("InspectDelegationCode() = %+v, want state %d delegate %s", got, test.want, test.delegate.Hex())
			}
		})
	}
}

func TestSignSetCodeAuthorization(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	authority := crypto.PubkeyToAddress(key.PublicKey)
	delegate := common.HexToAddress("0x1111111111111111111111111111111111111111")
	service := &ChainService{
		chainId:     big.NewInt(1),
		priKey:      key,
		fromAddress: authority.Hex(),
	}

	authorization, err := service.SignSetCodeAuthorization(delegate.Hex(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if authorization.Address != delegate || authorization.Nonce != 7 {
		t.Fatalf("unexpected authorization: %+v", authorization)
	}
	recovered, err := authorization.Authority()
	if err != nil {
		t.Fatal(err)
	}
	if recovered != authority {
		t.Fatalf("recovered authority %s, want %s", recovered.Hex(), authority.Hex())
	}
}

func TestPackCollectCalldata(t *testing.T) {
	items := []batchsweepexecutor.BatchSweepExecutorCollectItem{{
		TaskId:       big.NewInt(9),
		Account:      common.HexToAddress("0x1111111111111111111111111111111111111111"),
		Token:        common.Address{}, // native currency
		Recipient:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
		Amount:       big.NewInt(123),
		CallGasLimit: big.NewInt(50_000),
	}}

	data, err := PackCollectCalldata(items)
	if err != nil {
		t.Fatal(err)
	}
	contractABI, err := batchsweepexecutor.BatchSweepExecutorMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 4 || !bytes.Equal(data[:4], contractABI.Methods["collect"].ID) {
		t.Fatalf("unexpected collect selector: %x", data)
	}
}

func TestBuildSignedSetCodeTx(t *testing.T) {
	sponsorKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	authorityKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	chainID := big.NewInt(1)
	delegate := common.HexToAddress("0x1111111111111111111111111111111111111111")
	authorityService := &ChainService{chainId: chainID, priKey: authorityKey}
	authorization, err := authorityService.SignSetCodeAuthorization(delegate.Hex(), 3)
	if err != nil {
		t.Fatal(err)
	}

	sponsorService := &ChainService{chainId: chainID, priKey: sponsorKey}
	destination := common.HexToAddress("0x2222222222222222222222222222222222222222")
	tx, err := sponsorService.BuildSignedSetCodeTx(SetCodeTxRequest{
		Nonce:          5,
		GasTipCap:      big.NewInt(2),
		GasFeeCap:      big.NewInt(20),
		GasLimit:       300_000,
		To:             destination,
		Data:           []byte{1, 2, 3, 4},
		Authorizations: []types.SetCodeAuthorization{authorization},
	})
	if err != nil {
		t.Fatal(err)
	}
	if tx.Type() != types.SetCodeTxType {
		t.Fatalf("transaction type %d, want %d", tx.Type(), types.SetCodeTxType)
	}
	if tx.Nonce() != 5 || tx.Gas() != 300_000 || tx.To() == nil || *tx.To() != destination {
		t.Fatalf("unexpected transaction fields: nonce=%d gas=%d to=%v", tx.Nonce(), tx.Gas(), tx.To())
	}
	if len(tx.SetCodeAuthorizations()) != 1 {
		t.Fatalf("authorization count %d, want 1", len(tx.SetCodeAuthorizations()))
	}

	sender, err := types.Sender(types.LatestSignerForChainID(chainID), tx)
	if err != nil {
		t.Fatal(err)
	}
	wantSender := crypto.PubkeyToAddress(sponsorKey.PublicKey)
	if sender != wantSender {
		t.Fatalf("sender %s, want %s", sender.Hex(), wantSender.Hex())
	}
}

func TestBuildSignedSetCodeTxRejectsEmptyAuthorizationList(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	service := &ChainService{chainId: big.NewInt(1), priKey: key}
	_, err = service.BuildSignedSetCodeTx(SetCodeTxRequest{
		GasTipCap: big.NewInt(1),
		GasFeeCap: big.NewInt(2),
		GasLimit:  21_000,
		To:        common.HexToAddress("0x1111111111111111111111111111111111111111"),
	})
	if err == nil {
		t.Fatal("expected empty authorization list error")
	}
}

func TestBuildSignedDynamicFeeTx(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	chainID := big.NewInt(56)
	service := &ChainService{chainId: chainID, priKey: key}
	destination := common.HexToAddress("0x2222222222222222222222222222222222222222")
	tx, err := service.BuildSignedDynamicFeeTx(SetCodeTxRequest{
		Nonce:     8,
		GasTipCap: big.NewInt(2),
		GasFeeCap: big.NewInt(20),
		GasLimit:  180_000,
		To:        destination,
		Data:      []byte{4, 3, 2, 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if tx.Type() != types.DynamicFeeTxType || tx.Nonce() != 8 || tx.Gas() != 180_000 {
		t.Fatalf("unexpected dynamic-fee transaction: type=%d nonce=%d gas=%d", tx.Type(), tx.Nonce(), tx.Gas())
	}
	sender, err := types.Sender(types.LatestSignerForChainID(chainID), tx)
	if err != nil {
		t.Fatal(err)
	}
	if sender != crypto.PubkeyToAddress(key.PublicKey) {
		t.Fatalf("unexpected sender %s", sender.Hex())
	}
}

func TestBuildSignedDynamicFeeTxRejectsAuthorization(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	service := &ChainService{chainId: big.NewInt(1), priKey: key}
	_, err = service.BuildSignedDynamicFeeTx(SetCodeTxRequest{
		GasTipCap:      big.NewInt(1),
		GasFeeCap:      big.NewInt(2),
		GasLimit:       21_000,
		To:             common.HexToAddress("0x1111111111111111111111111111111111111111"),
		Authorizations: []types.SetCodeAuthorization{{}},
	})
	if err == nil {
		t.Fatal("expected authorization rejection")
	}
}

func TestSponsoredCallMsgOmitsEmptyAuthorizationList(t *testing.T) {
	request := SetCodeTxRequest{
		GasTipCap:      big.NewInt(1),
		GasFeeCap:      big.NewInt(2),
		To:             common.HexToAddress("0x1111111111111111111111111111111111111111"),
		Authorizations: make([]types.SetCodeAuthorization, 0),
	}

	callMsg := sponsoredCallMsg(common.Address{}, request, new(big.Int))
	if callMsg.AuthorizationList != nil {
		t.Fatalf("empty authorization list must be omitted, got %#v", callMsg.AuthorizationList)
	}

	request.Authorizations = []types.SetCodeAuthorization{{Nonce: 1}}
	callMsg = sponsoredCallMsg(common.Address{}, request, new(big.Int))
	if len(callMsg.AuthorizationList) != 1 {
		t.Fatalf("non-empty authorization list must be preserved, got %#v", callMsg.AuthorizationList)
	}
}

func TestCalculateSetCodeFeeCap(t *testing.T) {
	feeCap, err := CalculateSetCodeFeeCap(big.NewInt(100), big.NewInt(3))
	if err != nil {
		t.Fatal(err)
	}
	if feeCap.Cmp(big.NewInt(203)) != 0 {
		t.Fatalf("fee cap %s, want 203", feeCap)
	}
}

func TestAddGasMargin(t *testing.T) {
	gasLimit, err := AddGasMargin(100_001, 2_000)
	if err != nil {
		t.Fatal(err)
	}
	if gasLimit != 120_002 {
		t.Fatalf("gas limit %d, want 120002", gasLimit)
	}

	if _, err := AddGasMargin(100, 10_001); err == nil {
		t.Fatal("expected excessive margin error")
	}
}

func TestUnsupportedSetCodeTransactionIsDefinitelyNotBroadcast(t *testing.T) {
	if !isTxDefinitelyNotBroadcast(errors.New("unsupported transaction type: 0x04")) {
		t.Fatal("unsupported type-4 transaction must be treated as not broadcast")
	}
}
