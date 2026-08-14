package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/pkg/contracts/erc20"
	"github.com/jiajia556/tool-box/log"
	"github.com/shopspring/decimal"
)

type TxStatus string

const (
	TxStatusNotFound  TxStatus = "not_found"
	TxStatusPending   TxStatus = "pending"
	TxStatusMined     TxStatus = "mined"
	TxStatusConfirmed TxStatus = "confirmed"
	TxStatusFailed    TxStatus = "failed"
	TxStatusUnknown   TxStatus = "unknown"
)

func (s *ChainService) GetTxStatus(txHash string) (TxStatus, error) {
	if s == nil || s.rpcClient == nil {
		return "", errors.New("GetTxStatus: chain service not initialized")
	}

	if !common.IsHexHash(txHash) {
		return "", errors.New("GetTxStatus: invalid transaction hash")
	}

	hash := common.HexToHash(txHash)

	tx, isPending, err := s.rpcClient.TransactionByHash(context.Background(), hash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return TxStatusNotFound, nil
		}
		return "", fmt.Errorf("GetTxStatus: transaction by hash: %w", err)
	}

	if tx == nil {
		return TxStatusNotFound, nil
	}

	if isPending {
		return TxStatusPending, nil
	}

	receipt, err := s.rpcClient.TransactionReceipt(context.Background(), hash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			log.Debug("transaction found but receipt not found", "chainDbId", s.chainDbId, "txHash", txHash)
			return TxStatusUnknown, nil
		}
		return "", fmt.Errorf("GetTxStatus: transaction receipt: %w", err)
	}

	if receipt == nil {
		log.Debug("transaction found but receipt is nil", "chainDbId", s.chainDbId, "txHash", txHash)
		return TxStatusUnknown, nil
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return TxStatusFailed, nil
	}

	if s.safeConfirmations == 0 {
		return TxStatusConfirmed, nil
	}

	latest, err := s.rpcClient.BlockNumber(context.Background())
	if err != nil {
		return "", fmt.Errorf("GetTxStatus: block number: %w", err)
	}

	if receipt.BlockNumber == nil {
		return TxStatusPending, nil
	}

	minedBlock := receipt.BlockNumber.Uint64()

	if latest < minedBlock {
		return TxStatusPending, nil
	}

	confirmations := latest - minedBlock + 1
	if confirmations >= s.safeConfirmations {
		return TxStatusConfirmed, nil
	}

	return TxStatusMined, nil
}

func (s *ChainService) BalanceAt(address string) (decimal.Decimal, error) {
	if s == nil || s.rpcClient == nil {
		return decimal.Zero, errors.New("BalanceAt: chain service not initialized")
	}
	if !common.IsHexAddress(address) {
		return decimal.Zero, errors.New("BalanceAt: invalid address")
	}
	balance, err := s.rpcClient.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return decimal.Zero, fmt.Errorf("BalanceAt: rpc balance: %w", err)
	}
	return decimal.NewFromBigInt(balance, 0), nil
}

func (s *ChainService) BalanceOf(tokenAddress string, address string) (decimal.Decimal, error) {
	if s == nil || s.rpcClient == nil {
		return decimal.Zero, errors.New("BalanceOf: chain service not initialized")
	}
	if !common.IsHexAddress(tokenAddress) {
		return decimal.Zero, errors.New("BalanceOf: invalid token address")
	}
	if !common.IsHexAddress(address) {
		return decimal.Zero, errors.New("BalanceOf: invalid address")
	}
	instance, err := erc20.NewErc20(common.HexToAddress(tokenAddress), s.rpcClient)
	if err != nil {
		return decimal.Zero, fmt.Errorf("BalanceOf: new erc20: %w", err)
	}
	balance, err := instance.BalanceOf(nil, common.HexToAddress(address))
	if err != nil {
		return decimal.Zero, fmt.Errorf("BalanceOf: contract balance: %w", err)
	}
	return decimal.NewFromBigInt(balance, 0), nil
}

func (s *ChainService) IsContract(address string) (bool, error) {
	if s == nil || s.rpcClient == nil {
		return false, errors.New("IsContract: chain service not initialized")
	}
	if !common.IsHexAddress(address) {
		return false, errors.New("IsContract: invalid address")
	}
	code, err := s.rpcClient.CodeAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return false, fmt.Errorf("IsContract: code at: %w", err)
	}
	return len(code) > 0, nil
}

func (s *ChainService) SuggestGasPrice() (decimal.Decimal, error) {
	if s == nil || s.rpcClient == nil {
		return decimal.Zero, errors.New("SuggestGasPrice: chain service not initialized")
	}
	price, err := s.rpcClient.SuggestGasPrice(context.Background())
	if err != nil {
		return decimal.Zero, fmt.Errorf("SuggestGasPrice: suggest: %w", err)
	}
	return decimal.NewFromBigInt(price, 0), nil
}

func (s *ChainService) IsNonceOccupied(address string, nonce uint64) (bool, error) {
	if s == nil || s.rpcClient == nil {
		return false, errors.New("IsNonceOccupied: chain service not initialized")
	}
	if !common.IsHexAddress(address) {
		return false, errors.New("IsNonceOccupied: invalid address")
	}

	// pending nonce = 当前地址“下一个可用 nonce”
	// 若传入 nonce 小于 pending nonce，说明该 nonce 已经被使用（已上链或在 pending 池中）
	pendingNonce, err := s.rpcClient.PendingNonceAt(context.Background(), common.HexToAddress(address))
	if err != nil {
		return false, fmt.Errorf("IsNonceOccupied: pending nonce: %w", err)
	}

	return nonce < pendingNonce, nil
}

func (s *ChainService) GetGasUsedAndEffectiveGasPrice(hash string) (decimal.Decimal, decimal.Decimal, error) {
	receipt, err := s.rpcClient.TransactionReceipt(context.Background(), common.HexToHash(hash))
	if err != nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("GetGasUsedAndEffectiveGasPrice: transaction receipt: %w", err)
	}

	gasUsed := receipt.GasUsed
	effectiveGasPrice := receipt.EffectiveGasPrice

	return decimal.NewFromUint64(gasUsed), decimal.NewFromBigInt(effectiveGasPrice, 0), nil
}
