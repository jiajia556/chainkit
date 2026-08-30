package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
	"github.com/jiajia556/chainkit/pkg/contracts/batchsweepexecutor"
)

// DelegationState describes the code currently installed at an EIP-7702 authority.
type DelegationState uint8

const (
	DelegationNone DelegationState = iota
	DelegationExpected
	DelegationOther
	DelegationUnsupportedCode
)

type DelegationInfo struct {
	State    DelegationState
	Delegate common.Address
}

// SetCodeTxRequest contains the outer sponsor transaction fields. Authorization
// signatures must already have been produced by the deposit-address keys.
type SetCodeTxRequest struct {
	Nonce          uint64
	GasTipCap      *big.Int
	GasFeeCap      *big.Int
	GasLimit       uint64
	To             common.Address
	Value          *big.Int
	Data           []byte
	AccessList     types.AccessList
	Authorizations []types.SetCodeAuthorization
}

const defaultSetCodeGasMarginBPS uint64 = 2_000 // 20%

// InspectDelegationCode classifies account code without an RPC dependency.
func InspectDelegationCode(code []byte, expectedDelegate common.Address) DelegationInfo {
	if len(code) == 0 {
		return DelegationInfo{State: DelegationNone}
	}
	delegate, ok := types.ParseDelegation(code)
	if !ok {
		return DelegationInfo{State: DelegationUnsupportedCode}
	}
	if delegate == expectedDelegate {
		return DelegationInfo{State: DelegationExpected, Delegate: delegate}
	}
	return DelegationInfo{State: DelegationOther, Delegate: delegate}
}

func (s *ChainService) DelegationAt(
	ctx context.Context,
	authority string,
	expectedDelegate string,
) (DelegationInfo, error) {
	if s == nil || s.rpcClient == nil {
		return DelegationInfo{}, errors.New("chain service not initialized")
	}
	if !common.IsHexAddress(authority) {
		return DelegationInfo{}, errors.New("invalid authority address")
	}
	if !common.IsHexAddress(expectedDelegate) {
		return DelegationInfo{}, errors.New("invalid expected delegate address")
	}

	code, err := s.rpcClient.CodeAt(ctx, common.HexToAddress(authority), nil)
	if err != nil {
		return DelegationInfo{}, err
	}
	return InspectDelegationCode(code, common.HexToAddress(expectedDelegate)), nil
}

// StableAuthorizationNonce returns the latest state nonce only when there is no
// pending nonce gap. Signing while the authority has pending transactions makes
// inclusion order ambiguous and can invalidate the EIP-7702 authorization.
func (s *ChainService) StableAuthorizationNonce(ctx context.Context, authority string) (uint64, error) {
	if s == nil || s.rpcClient == nil {
		return 0, errors.New("chain service not initialized")
	}
	if !common.IsHexAddress(authority) {
		return 0, errors.New("invalid authority address")
	}
	address := common.HexToAddress(authority)

	latestNonce, err := s.rpcClient.NonceAt(ctx, address, nil)
	if err != nil {
		return 0, fmt.Errorf("read latest authority nonce: %w", err)
	}
	pendingNonce, err := s.rpcClient.PendingNonceAt(ctx, address)
	if err != nil {
		return 0, fmt.Errorf("read pending authority nonce: %w", err)
	}
	if latestNonce != pendingNonce {
		return 0, fmt.Errorf(
			"authority has pending transactions: latest nonce %d, pending nonce %d",
			latestNonce,
			pendingNonce,
		)
	}
	return latestNonce, nil
}

// SignSetCodeAuthorization signs an authorization using the address currently
// selected on ChainService (normally a user deposit address).
func (s *ChainService) SignSetCodeAuthorization(
	delegate string,
	nonce uint64,
) (types.SetCodeAuthorization, error) {
	if s == nil || s.priKey == nil || s.chainId == nil {
		return types.SetCodeAuthorization{}, errors.New("chain service signer not initialized")
	}
	if !common.IsHexAddress(delegate) || common.HexToAddress(delegate) == (common.Address{}) {
		return types.SetCodeAuthorization{}, errors.New("invalid delegate address")
	}
	chainID, overflow := uint256.FromBig(s.chainId)
	if overflow || s.chainId.Sign() <= 0 {
		return types.SetCodeAuthorization{}, errors.New("invalid chain id")
	}

	authorization, err := types.SignSetCode(s.priKey, types.SetCodeAuthorization{
		ChainID: *chainID,
		Address: common.HexToAddress(delegate),
		Nonce:   nonce,
	})
	if err != nil {
		return types.SetCodeAuthorization{}, err
	}

	authority, err := authorization.Authority()
	if err != nil {
		return types.SetCodeAuthorization{}, fmt.Errorf("recover signed authority: %w", err)
	}
	if s.fromAddress != "" && authority != common.HexToAddress(s.fromAddress) {
		return types.SetCodeAuthorization{}, errors.New("authorization signer does not match selected address")
	}
	return authorization, nil
}

// PackCollectCalldata ABI-encodes BatchSweepExecutor.collect(items).
func PackCollectCalldata(items []batchsweepexecutor.BatchSweepExecutorCollectItem) ([]byte, error) {
	if len(items) == 0 {
		return nil, errors.New("collect items are empty")
	}
	contractABI, err := batchsweepexecutor.BatchSweepExecutorMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("parse batch sweep executor ABI: %w", err)
	}
	data, err := contractABI.Pack("collect", items)
	if err != nil {
		return nil, fmt.Errorf("pack collect calldata: %w", err)
	}
	return data, nil
}

// SuggestSetCodeFees returns EIP-1559 fee caps suitable for a type-4 transaction.
// The fee cap allows the next block's base fee to rise up to roughly 2x while
// preserving the suggested priority fee.
func (s *ChainService) SuggestSetCodeFees(ctx context.Context) (gasTipCap, gasFeeCap *big.Int, err error) {
	if s == nil || s.rpcClient == nil {
		return nil, nil, errors.New("chain service not initialized")
	}

	gasTipCap, err = s.rpcClient.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("suggest gas tip cap: %w", err)
	}
	header, err := s.rpcClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("read latest block header: %w", err)
	}
	if header == nil || header.BaseFee == nil {
		return nil, nil, errors.New("latest block does not contain an EIP-1559 base fee")
	}
	gasFeeCap, err = CalculateSetCodeFeeCap(header.BaseFee, gasTipCap)
	if err != nil {
		return nil, nil, err
	}
	return gasTipCap, gasFeeCap, nil
}

func CalculateSetCodeFeeCap(baseFee, gasTipCap *big.Int) (*big.Int, error) {
	if baseFee == nil || gasTipCap == nil {
		return nil, errors.New("base fee and gas tip cap are required")
	}
	if baseFee.Sign() < 0 || gasTipCap.Sign() < 0 {
		return nil, errors.New("base fee and gas tip cap cannot be negative")
	}
	feeCap := new(big.Int).Mul(baseFee, big.NewInt(2))
	feeCap.Add(feeCap, gasTipCap)
	if _, overflow := uint256.FromBig(feeCap); overflow {
		return nil, errors.New("calculated fee cap overflows uint256")
	}
	return feeCap, nil
}

// EstimateSetCodeGas estimates the complete sponsored call, including processing
// the authorization list and executing BatchSweepExecutor calldata.
func (s *ChainService) EstimateSetCodeGas(
	ctx context.Context,
	request SetCodeTxRequest,
) (uint64, error) {
	if len(request.Authorizations) == 0 {
		return 0, errors.New("set-code transaction authorization list is empty")
	}
	return s.estimateSponsoredGas(ctx, request)
}

// EstimateDynamicFeeGas estimates a recurring sponsored collect call after all
// authorities have already delegated to the expected implementation.
func (s *ChainService) EstimateDynamicFeeGas(
	ctx context.Context,
	request SetCodeTxRequest,
) (uint64, error) {
	if len(request.Authorizations) != 0 {
		return 0, errors.New("dynamic-fee transaction cannot contain authorizations")
	}
	return s.estimateSponsoredGas(ctx, request)
}

func (s *ChainService) estimateSponsoredGas(
	ctx context.Context,
	request SetCodeTxRequest,
) (uint64, error) {
	if s == nil || s.rpcClient == nil || s.priKey == nil {
		return 0, errors.New("chain service sponsor not initialized")
	}
	if request.To == (common.Address{}) {
		return 0, errors.New("sponsored transaction destination is zero")
	}
	if err := validateFeeCaps(request.GasTipCap, request.GasFeeCap); err != nil {
		return 0, err
	}
	value := request.Value
	if value == nil {
		value = new(big.Int)
	}
	if value.Sign() < 0 {
		return 0, errors.New("sponsored transaction value cannot be negative")
	}

	from := crypto.PubkeyToAddress(s.priKey.PublicKey)
	estimated, err := s.rpcClient.EstimateGas(ctx, ethereum.CallMsg{
		From:              from,
		To:                &request.To,
		GasFeeCap:         request.GasFeeCap,
		GasTipCap:         request.GasTipCap,
		Value:             value,
		Data:              request.Data,
		AccessList:        request.AccessList,
		AuthorizationList: request.Authorizations,
	})
	if err != nil {
		return 0, fmt.Errorf("estimate sponsored transaction gas: %w", err)
	}
	gasLimit, err := AddGasMargin(estimated, defaultSetCodeGasMarginBPS)
	if err != nil {
		return 0, err
	}
	return gasLimit, nil
}

// BuildSignedDynamicFeeTx builds a type-2 sponsored call for authorities that
// are already delegated. It deliberately rejects authorization entries.
func (s *ChainService) BuildSignedDynamicFeeTx(request SetCodeTxRequest) (*types.Transaction, error) {
	if s == nil || s.priKey == nil || s.chainId == nil {
		return nil, errors.New("chain service sponsor signer not initialized")
	}
	if request.To == (common.Address{}) {
		return nil, errors.New("dynamic-fee transaction destination is zero")
	}
	if request.GasLimit == 0 {
		return nil, errors.New("dynamic-fee transaction gas limit is zero")
	}
	if len(request.Authorizations) != 0 {
		return nil, errors.New("dynamic-fee transaction cannot contain authorizations")
	}
	if err := validateFeeCaps(request.GasTipCap, request.GasFeeCap); err != nil {
		return nil, err
	}
	value := request.Value
	if value == nil {
		value = new(big.Int)
	}
	if value.Sign() < 0 {
		return nil, errors.New("invalid dynamic-fee transaction value")
	}

	unsignedTx := types.NewTx(&types.DynamicFeeTx{
		ChainID:    new(big.Int).Set(s.chainId),
		Nonce:      request.Nonce,
		GasTipCap:  new(big.Int).Set(request.GasTipCap),
		GasFeeCap:  new(big.Int).Set(request.GasFeeCap),
		Gas:        request.GasLimit,
		To:         &request.To,
		Value:      new(big.Int).Set(value),
		Data:       common.CopyBytes(request.Data),
		AccessList: append(types.AccessList(nil), request.AccessList...),
	})
	signedTx, err := types.SignTx(unsignedTx, types.LatestSignerForChainID(s.chainId), s.priKey)
	if err != nil {
		return nil, fmt.Errorf("sign dynamic-fee transaction: %w", err)
	}
	return signedTx, nil
}

// AddGasMargin applies a basis-point margin and rounds up.
func AddGasMargin(estimated uint64, marginBPS uint64) (uint64, error) {
	if estimated == 0 {
		return 0, errors.New("estimated gas is zero")
	}
	if marginBPS > 10_000 {
		return 0, errors.New("gas margin cannot exceed 100 percent")
	}
	value := new(big.Int).SetUint64(estimated)
	value.Mul(value, new(big.Int).SetUint64(10_000+marginBPS))
	value.Add(value, big.NewInt(9_999))
	value.Div(value, big.NewInt(10_000))
	if !value.IsUint64() {
		return 0, errors.New("gas limit overflows uint64")
	}
	return value.Uint64(), nil
}

// BuildSignedSetCodeTx builds and signs the type-4 outer transaction with the
// address currently selected on ChainService acting as the gas sponsor.
func (s *ChainService) BuildSignedSetCodeTx(request SetCodeTxRequest) (*types.Transaction, error) {
	if s == nil || s.priKey == nil || s.chainId == nil {
		return nil, errors.New("chain service sponsor signer not initialized")
	}
	if request.To == (common.Address{}) {
		return nil, errors.New("set-code transaction destination is zero")
	}
	if request.GasLimit == 0 {
		return nil, errors.New("set-code transaction gas limit is zero")
	}
	if len(request.Authorizations) == 0 {
		return nil, errors.New("set-code transaction authorization list is empty")
	}
	if err := validateFeeCaps(request.GasTipCap, request.GasFeeCap); err != nil {
		return nil, err
	}

	chainID, overflow := uint256.FromBig(s.chainId)
	if overflow || s.chainId.Sign() <= 0 {
		return nil, errors.New("invalid chain id")
	}
	value := request.Value
	if value == nil {
		value = new(big.Int)
	}
	value256, overflow := uint256.FromBig(value)
	if overflow || value.Sign() < 0 {
		return nil, errors.New("invalid set-code transaction value")
	}
	tip256, _ := uint256.FromBig(request.GasTipCap)
	fee256, _ := uint256.FromBig(request.GasFeeCap)

	seenAuthorities := make(map[common.Address]struct{}, len(request.Authorizations))
	for i := range request.Authorizations {
		authorization := &request.Authorizations[i]
		if authorization.ChainID.ToBig().Cmp(s.chainId) != 0 {
			return nil, fmt.Errorf("authorization %d has a different chain id", i)
		}
		authority, err := authorization.Authority()
		if err != nil {
			return nil, fmt.Errorf("authorization %d has an invalid signature: %w", i, err)
		}
		if _, exists := seenAuthorities[authority]; exists {
			return nil, fmt.Errorf("authorization %d duplicates authority %s", i, authority.Hex())
		}
		seenAuthorities[authority] = struct{}{}
	}

	unsignedTx := types.NewTx(&types.SetCodeTx{
		ChainID:    chainID,
		Nonce:      request.Nonce,
		GasTipCap:  tip256,
		GasFeeCap:  fee256,
		Gas:        request.GasLimit,
		To:         request.To,
		Value:      value256,
		Data:       common.CopyBytes(request.Data),
		AccessList: request.AccessList,
		AuthList:   append([]types.SetCodeAuthorization(nil), request.Authorizations...),
	})

	signedTx, err := types.SignTx(
		unsignedTx,
		types.LatestSignerForChainID(s.chainId),
		s.priKey,
	)
	if err != nil {
		return nil, fmt.Errorf("sign set-code transaction: %w", err)
	}
	return signedTx, nil
}

// SendSetCodeTx signs and broadcasts a type-4 transaction. A non-nil fakeErr
// means the RPC result was uncertain after a signed transaction hash existed;
// callers must track the hash instead of immediately retrying with another nonce.
func (s *ChainService) SendSetCodeTx(
	ctx context.Context,
	request SetCodeTxRequest,
) (hash string, fakeErr, err error) {
	if s == nil || s.rpcClient == nil {
		return "", nil, errors.New("chain service not initialized")
	}
	signedTx, err := s.BuildSignedSetCodeTx(request)
	if err != nil {
		return "", nil, err
	}
	return s.SendSignedTransaction(ctx, signedTx)
}

// SendSignedTransaction broadcasts an already signed transaction. It is kept
// separate from building so callers can durably persist rawTx and hash first.
func (s *ChainService) SendSignedTransaction(
	ctx context.Context,
	signedTx *types.Transaction,
) (hash string, fakeErr, err error) {
	if s == nil || s.rpcClient == nil {
		return "", nil, errors.New("chain service not initialized")
	}
	if signedTx == nil {
		return "", nil, errors.New("signed transaction is nil")
	}

	err = s.rpcClient.SendTransaction(ctx, signedTx)
	if err == nil {
		return signedTx.Hash().Hex(), nil, nil
	}
	if isTxDefinitelyNotBroadcast(err) {
		return "", nil, err
	}
	hash = signedTx.Hash().Hex()
	return hash, uncertainBroadcastError(hash, err), nil
}

func validateFeeCaps(gasTipCap, gasFeeCap *big.Int) error {
	if gasTipCap == nil || gasFeeCap == nil {
		return errors.New("set-code transaction fee caps are required")
	}
	if gasTipCap.Sign() < 0 || gasFeeCap.Sign() < 0 {
		return errors.New("set-code transaction fee caps cannot be negative")
	}
	if gasFeeCap.Cmp(gasTipCap) < 0 {
		return errors.New("set-code transaction fee cap is lower than tip cap")
	}
	if _, overflow := uint256.FromBig(gasTipCap); overflow {
		return errors.New("set-code transaction tip cap overflows uint256")
	}
	if _, overflow := uint256.FromBig(gasFeeCap); overflow {
		return errors.New("set-code transaction fee cap overflows uint256")
	}
	return nil
}
