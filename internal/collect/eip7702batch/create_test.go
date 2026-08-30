package eip7702batch

import (
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
)

const (
	testSponsor   = "0x1111111111111111111111111111111111111111"
	testExecutor  = "0x2222222222222222222222222222222222222222"
	testAuthority = "0x3333333333333333333333333333333333333333"
	testDelegate  = "0x4444444444444444444444444444444444444444"
	testRecipient = "0x5555555555555555555555555555555555555555"
	testToken     = "0x6666666666666666666666666666666666666666"
)

func TestPrepareSetCodeBatch(t *testing.T) {
	prepared, err := prepare(validRequest())
	if err != nil {
		t.Fatal(err)
	}
	if prepared.batch.AuthorizationCount != 1 {
		t.Fatalf("authorization count %d, want 1", prepared.batch.AuthorizationCount)
	}
	if len(prepared.items) != 1 || prepared.items[0].ItemIndex != 0 {
		t.Fatalf("unexpected prepared items: %+v", prepared.items)
	}
	if prepared.items[0].PlannedAmount != "123" {
		t.Fatalf("planned amount %q, want 123", prepared.items[0].PlannedAmount)
	}
}

func TestPrepareRejectsDuplicateTask(t *testing.T) {
	request := validRequest()
	request.Items = append(request.Items, request.Items[0])
	if _, err := prepare(request); err == nil {
		t.Fatal("expected duplicate task error")
	}
}

func TestPrepareRejectsType4WithoutAuthorization(t *testing.T) {
	request := validRequest()
	request.Items[0].AuthorizationRequired = false
	if _, err := prepare(request); err == nil {
		t.Fatal("expected missing authorization error")
	}
}

func TestPrepareAcceptsNativeItem(t *testing.T) {
	request := validRequest()
	request.Items[0].TokenAddress = "0x0000000000000000000000000000000000000000"
	if _, err := prepare(request); err != nil {
		t.Fatal(err)
	}
}

func validRequest() CreateRequest {
	return CreateRequest{
		ChainDbId:                1,
		SponsorMnemonicAddressId: 2,
		SponsorAddress:           testSponsor,
		SponsorNonce:             0,
		ExecutorAddress:          testExecutor,
		TxType:                   types.SetCodeTxType,
		Items: []ItemRequest{{
			CollectTaskId:         10,
			UserDepositAddressId:  11,
			AuthorityAddress:      testAuthority,
			AuthorizationRequired: true,
			AuthorizationNonce:    0,
			DelegateAddress:       testDelegate,
			TokenAddress:          testToken,
			RecipientAddress:      testRecipient,
			PlannedAmount:         "123",
			CallGasLimit:          100_000,
		}},
	}
}
