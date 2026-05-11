package service

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jiajia556/chainkit/pkg/contracts/erc20"
	"github.com/shopspring/decimal"
)

func (s *ChainService) ApproveERC20(tokenAddress, spenderAddress string, amount decimal.Decimal, opts ...Option) (string, error) {
	if s == nil || s.client == nil {
		return "", errors.New("chain service not initialized")
	}
	if s.fromAddress == "" {
		return "", errors.New("chain service from not set")
	}
	if !common.IsHexAddress(tokenAddress) {
		return "", errors.New("invalid token address")
	}
	if !common.IsHexAddress(spenderAddress) {
		return "", errors.New("invalid address")
	}

	instance, err := erc20.NewErc20(common.HexToAddress(tokenAddress), s.client)
	if err != nil {
		return "", err
	}
	txOpts, err := s.GetBindTransactOpts(opts...)
	if err != nil {
		return "", err
	}
	tx, err := instance.Approve(txOpts, common.HexToAddress(spenderAddress), amount.BigInt())
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}
