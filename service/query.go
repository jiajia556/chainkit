package service

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jiajia556/chainkit/pkg/contracts/erc20"
	"github.com/shopspring/decimal"
)

type TxStatus string

const (
	TxStatusPending   TxStatus = "pending"
	TxStatusConfirmed TxStatus = "confirmed"
	TxStatusFailed    TxStatus = "failed"
)

func (s *ChainService) GetTxStatus(txHash string) (TxStatus, error) {
	if s == nil || s.client == nil {
		return "", errors.New("chain service not initialized")
	}
	if !common.IsHexHash(txHash) {
		return "", errors.New("invalid transaction hash")
	}
	_, isPending, err := s.client.TransactionByHash(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return "", err
	}
	if isPending {
		return TxStatusPending, nil
	}
	receipt, err := s.client.TransactionReceipt(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return "", err
	}
	if receipt.Status == 1 {
		return TxStatusConfirmed, nil
	}
	return TxStatusFailed, nil
}

func (s *ChainService) BalanceAt(address string) (decimal.Decimal, error) {
	if s == nil || s.client == nil {
		return decimal.Zero, errors.New("chain service not initialized")
	}
	if !common.IsHexAddress(address) {
		return decimal.Zero, errors.New("invalid address")
	}
	balance, err := s.client.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromBigInt(balance, 0), nil
}

func (s *ChainService) BalanceOf(tokenAddress string, address string) (decimal.Decimal, error) {
	if s == nil || s.client == nil {
		return decimal.Zero, errors.New("chain service not initialized")
	}
	if !common.IsHexAddress(tokenAddress) {
		return decimal.Zero, errors.New("invalid token address")
	}
	if !common.IsHexAddress(address) {
		return decimal.Zero, errors.New("invalid address")
	}
	instance, err := erc20.NewErc20(common.HexToAddress(tokenAddress), s.client)
	if err != nil {
		return decimal.Zero, err
	}
	balance, err := instance.BalanceOf(nil, common.HexToAddress(address))
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromBigInt(balance, 0), nil
}
