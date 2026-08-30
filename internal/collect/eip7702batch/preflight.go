package eip7702batch

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jiajia556/chainkit/models/chainkitcollectbatchitems"
	"github.com/jiajia556/chainkit/pkg/contracts/batchsweepexecutor"
	"github.com/jiajia556/chainkit/service"
)

func validateExecutorForBatch(
	ctx context.Context,
	chainService *service.ChainService,
	executorAddress string,
	sponsorAddress string,
	items *chainkitcollectbatchitems.List,
) error {
	if chainService == nil || chainService.GetClient() == nil {
		return errors.New("chain service is not initialized")
	}
	if !validNonZeroAddress(executorAddress) || !validNonZeroAddress(sponsorAddress) {
		return errors.New("executor or sponsor address is invalid")
	}
	if items == nil || items.IsEmpty() {
		return errors.New("collect batch has no items")
	}

	executor, err := batchsweepexecutor.NewBatchSweepExecutor(
		common.HexToAddress(executorAddress),
		chainService.GetClient(),
	)
	if err != nil {
		return fmt.Errorf("bind batch sweep executor: %w", err)
	}
	callOpts := &bind.CallOpts{Context: ctx}
	operator, err := executor.Operator(callOpts)
	if err != nil {
		return fmt.Errorf("read batch sweep operator: %w", err)
	}
	if operator != common.HexToAddress(sponsorAddress) {
		return fmt.Errorf("batch sweep operator %s does not match sponsor %s", operator.Hex(), common.HexToAddress(sponsorAddress).Hex())
	}
	paused, err := executor.Paused(callOpts)
	if err != nil {
		return fmt.Errorf("read batch sweep pause state: %w", err)
	}
	if paused {
		return errors.New("batch sweep executor is paused")
	}
	delegate, err := executor.SweepDelegate(callOpts)
	if err != nil {
		return fmt.Errorf("read batch sweep delegate: %w", err)
	}
	if delegate == (common.Address{}) {
		return errors.New("batch sweep delegate is not initialized")
	}
	maximum, err := executor.MAXBATCHITEMS(callOpts)
	if err != nil {
		return fmt.Errorf("read batch sweep maximum: %w", err)
	}
	if !maximum.IsUint64() || uint64(len(*items.Records)) > maximum.Uint64() {
		return fmt.Errorf("batch item count %d exceeds on-chain maximum %s", len(*items.Records), maximum.String())
	}

	recipients := make(map[common.Address]struct{})
	for _, item := range *items.Records {
		if common.HexToAddress(item.DelegateAddress) != delegate {
			return fmt.Errorf("batch item %d delegate does not match executor delegate %s", item.Id, delegate.Hex())
		}
		recipient := common.HexToAddress(item.RecipientAddress)
		if _, checked := recipients[recipient]; checked {
			continue
		}
		allowed, err := executor.AllowedRecipient(callOpts, recipient)
		if err != nil {
			return fmt.Errorf("read recipient permission for %s: %w", recipient.Hex(), err)
		}
		if !allowed {
			return fmt.Errorf("recipient %s is not allowed by batch sweep executor", recipient.Hex())
		}
		recipients[recipient] = struct{}{}
	}
	return nil
}
