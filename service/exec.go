package service

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/pkg/contracts/erc20"
	"github.com/shopspring/decimal"
)

func (s *ChainService) ApproveERC20(tokenAddress, spenderAddress string, amount decimal.Decimal, opts ...Option) (hash string, fakeErr, err error) {
	if s == nil || s.rpcClient == nil {
		return "", nil, errors.New("chain service not initialized")
	}
	if s.fromAddress == "" {
		return "", nil, errors.New("chain service from not set")
	}
	if !common.IsHexAddress(tokenAddress) {
		return "", nil, errors.New("invalid token address")
	}
	if !common.IsHexAddress(spenderAddress) {
		return "", nil, errors.New("invalid address")
	}

	instance, err := erc20.NewErc20(common.HexToAddress(tokenAddress), s.rpcClient)
	if err != nil {
		return "", nil, err
	}
	txOpts, err := s.GetBindTransactOpts(opts...)
	if err != nil {
		return "", nil, err
	}

	var lastSignedTx *types.Transaction
	originalSigner := txOpts.Signer
	txOpts.Signer = func(addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
		signedTx, signErr := originalSigner(addr, tx)
		if signErr != nil {
			return nil, signErr
		}
		lastSignedTx = signedTx
		return signedTx, nil
	}

	tx, err := instance.Approve(txOpts, common.HexToAddress(spenderAddress), amount.BigInt())
	if err != nil {
		if lastSignedTx == nil {
			return "", nil, err
		}
		hash = lastSignedTx.Hash().Hex()
		fakeErr = err
		return hash, fakeErr, nil
	}
	return tx.Hash().Hex(), nil, nil
}
