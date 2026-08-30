package eip7702batch

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/models/chainkitcollectbatches"
	"github.com/jiajia556/chainkit/models/chainkitcollectbatchitems"
	"github.com/jiajia556/chainkit/models/chainkitcollecttasks"
	"github.com/jiajia556/chainkit/pkg/contracts/batchsweepexecutor"
	"github.com/jiajia556/chainkit/service"
)

// Execute claims, signs and broadcasts one waiting sponsored collect batch.
// Type-4 installs delegation for first-time authorities; type-2 handles later
// collections from authorities that are already delegated. The signed raw
// transaction is persisted before the RPC broadcast starts.
func Execute(
	ctx context.Context,
	batchId uint64,
	depositPassword string,
	sponsorPassword string,
) error {
	if batchId == 0 {
		return errors.New("batch id is zero")
	}

	batch := chainkitcollectbatches.NewRecord()
	if err := batch.Read(batchId); err != nil {
		return fmt.Errorf("read collect batch: %w", err)
	}
	if batch.Model.TxType != types.SetCodeTxType && batch.Model.TxType != types.DynamicFeeTxType {
		return fmt.Errorf("batch %d has unsupported transaction type %d", batchId, batch.Model.TxType)
	}
	claimed, err := batch.ClaimWaiting()
	if err != nil {
		return fmt.Errorf("claim collect batch: %w", err)
	}
	if !claimed {
		return nil
	}

	failWaiting := func(cause error) error {
		if updateErr := batch.SetWaitingWithError(cause.Error()); updateErr != nil {
			return fmt.Errorf("%v; reset batch to waiting: %w", cause, updateErr)
		}
		return cause
	}

	itemList := chainkitcollectbatchitems.NewList()
	itemList.GetByBatchId(batchId)
	if itemList.IsEmpty() {
		return failWaiting(errors.New("collect batch has no items"))
	}

	sponsorService, err := service.NewChainService(batch.Model.ChainDbId)
	if err != nil {
		return failWaiting(fmt.Errorf("create sponsor chain service: %w", err))
	}
	defer sponsorService.CloseClient()
	if err := sponsorService.SetFromByMnemonicAddress(
		batch.Model.SponsorMnemonicAddressId,
		sponsorPassword,
	); err != nil {
		return failWaiting(fmt.Errorf("set sponsor signer: %w", err))
	}
	sponsorAddress, err := sponsorService.GetFromAddress()
	if err != nil {
		return failWaiting(err)
	}
	if common.HexToAddress(sponsorAddress) != common.HexToAddress(batch.Model.SponsorAddress) {
		return failWaiting(errors.New("configured sponsor does not match batch sponsor address"))
	}
	liveSponsorNonce, err := sponsorService.GetClient().PendingNonceAt(
		ctx,
		common.HexToAddress(sponsorAddress),
	)
	if err != nil {
		return failWaiting(fmt.Errorf("read sponsor nonce: %w", err))
	}
	if liveSponsorNonce != batch.Model.SponsorNonce {
		updated, updateErr := batch.RefreshSponsorNonce(liveSponsorNonce)
		if updateErr != nil {
			return failWaiting(fmt.Errorf("refresh sponsor nonce: %w", updateErr))
		}
		if !updated {
			return errors.New("batch changed before sponsor nonce could be refreshed")
		}
	}
	if err := validateExecutorForBatch(
		ctx,
		sponsorService,
		batch.Model.ExecutorAddress,
		sponsorAddress,
		itemList,
	); err != nil {
		return failWaiting(fmt.Errorf("validate batch sweep executor: %w", err))
	}

	collectItems := make([]batchsweepexecutor.BatchSweepExecutorCollectItem, 0, len(*itemList.Records))
	authorizations := make([]types.SetCodeAuthorization, 0, batch.Model.AuthorizationCount)
	signedAuthorities := make(map[common.Address]struct{}, batch.Model.AuthorizationCount)

	for _, item := range *itemList.Records {
		amount, ok := new(big.Int).SetString(item.PlannedAmount, 10)
		if !ok || amount.Sign() <= 0 {
			return failWaiting(fmt.Errorf("batch item %d has invalid planned amount", item.Id))
		}
		collectItems = append(collectItems, batchsweepexecutor.BatchSweepExecutorCollectItem{
			TaskId:       new(big.Int).SetUint64(item.CollectTaskId),
			Account:      common.HexToAddress(item.AuthorityAddress),
			Token:        common.HexToAddress(item.TokenAddress),
			Recipient:    common.HexToAddress(item.RecipientAddress),
			Amount:       amount,
			CallGasLimit: new(big.Int).SetUint64(item.CallGasLimit),
		})

		if !item.AuthorizationRequired {
			delegation, err := sponsorService.DelegationAt(
				ctx,
				item.AuthorityAddress,
				item.DelegateAddress,
			)
			if err != nil {
				return failWaiting(fmt.Errorf("inspect delegated authority %s: %w", item.AuthorityAddress, err))
			}
			if delegation.State != service.DelegationExpected {
				return failWaiting(fmt.Errorf("authority %s is no longer delegated to the expected implementation", item.AuthorityAddress))
			}
			continue
		}

		authority := common.HexToAddress(item.AuthorityAddress)
		if _, exists := signedAuthorities[authority]; exists {
			continue
		}

		delegation, err := sponsorService.DelegationAt(
			ctx,
			item.AuthorityAddress,
			item.DelegateAddress,
		)
		if err != nil {
			return failWaiting(fmt.Errorf("inspect authority %s: %w", item.AuthorityAddress, err))
		}
		if delegation.State != service.DelegationNone {
			return failWaiting(fmt.Errorf("authority %s delegation state changed before signing", item.AuthorityAddress))
		}

		authorityService, err := service.NewChainService(batch.Model.ChainDbId)
		if err != nil {
			return failWaiting(fmt.Errorf("create authority chain service: %w", err))
		}
		if err := authorityService.SetFromByDepositAddress(item.UserDepositAddressId, depositPassword); err != nil {
			authorityService.CloseClient()
			return failWaiting(fmt.Errorf("set authority signer %s: %w", item.AuthorityAddress, err))
		}
		selectedAddress, err := authorityService.GetFromAddress()
		if err != nil {
			authorityService.CloseClient()
			return failWaiting(err)
		}
		if common.HexToAddress(selectedAddress) != authority {
			authorityService.CloseClient()
			return failWaiting(fmt.Errorf("deposit address id %d does not match authority %s", item.UserDepositAddressId, authority.Hex()))
		}
		liveNonce, err := authorityService.StableAuthorizationNonce(ctx, item.AuthorityAddress)
		if err != nil {
			authorityService.CloseClient()
			return failWaiting(fmt.Errorf("read stable authorization nonce for %s: %w", item.AuthorityAddress, err))
		}
		if liveNonce != item.AuthorizationNonce {
			authorityService.CloseClient()
			return failWaiting(fmt.Errorf(
				"authority %s nonce changed: stored %d, live %d",
				item.AuthorityAddress,
				item.AuthorizationNonce,
				liveNonce,
			))
		}
		authorization, err := authorityService.SignSetCodeAuthorization(
			item.DelegateAddress,
			item.AuthorizationNonce,
		)
		authorityService.CloseClient()
		if err != nil {
			return failWaiting(fmt.Errorf("sign authorization for %s: %w", item.AuthorityAddress, err))
		}
		authorizations = append(authorizations, authorization)
		signedAuthorities[authority] = struct{}{}
	}

	if batch.Model.TxType == types.SetCodeTxType && len(authorizations) == 0 {
		return failWaiting(errors.New("type-4 batch has no authorizations after preflight"))
	}
	if batch.Model.TxType == types.DynamicFeeTxType && len(authorizations) != 0 {
		return failWaiting(errors.New("type-2 batch unexpectedly contains authorizations"))
	}
	if uint32(len(authorizations)) != batch.Model.AuthorizationCount {
		return failWaiting(fmt.Errorf(
			"authorization count changed: batch %d, built %d",
			batch.Model.AuthorizationCount,
			len(authorizations),
		))
	}

	calldata, err := service.PackCollectCalldata(collectItems)
	if err != nil {
		return failWaiting(err)
	}
	gasTipCap, gasFeeCap, err := sponsorService.SuggestSetCodeFees(ctx)
	if err != nil {
		return failWaiting(err)
	}
	request := service.SetCodeTxRequest{
		Nonce:          batch.Model.SponsorNonce,
		GasTipCap:      gasTipCap,
		GasFeeCap:      gasFeeCap,
		To:             common.HexToAddress(batch.Model.ExecutorAddress),
		Data:           calldata,
		Authorizations: authorizations,
	}
	if batch.Model.TxType == types.SetCodeTxType {
		request.GasLimit, err = sponsorService.EstimateSetCodeGas(ctx, request)
	} else {
		request.GasLimit, err = sponsorService.EstimateDynamicFeeGas(ctx, request)
	}
	if err != nil {
		return failWaiting(err)
	}

	var signedTx *types.Transaction
	if batch.Model.TxType == types.SetCodeTxType {
		signedTx, err = sponsorService.BuildSignedSetCodeTx(request)
	} else {
		signedTx, err = sponsorService.BuildSignedDynamicFeeTx(request)
	}
	if err != nil {
		return failWaiting(err)
	}
	rawTx, err := signedTx.MarshalBinary()
	if err != nil {
		return failWaiting(fmt.Errorf("encode signed set-code transaction: %w", err))
	}
	built, err := batch.SetBuilt(
		signedTx.Hash().Hex(),
		rawTx,
		request.GasLimit,
		gasFeeCap.String(),
		gasTipCap.String(),
		uint32(len(authorizations)),
	)
	if err != nil {
		return fmt.Errorf("persist signed collect batch: %w", err)
	}
	if !built {
		return errors.New("collect batch state changed before signed transaction was persisted")
	}

	hash, fakeErr, sendErr := sponsorService.SendSignedTransaction(ctx, signedTx)
	if sendErr != nil {
		return failWaiting(fmt.Errorf("broadcast collect batch: %w", sendErr))
	}
	if fakeErr != nil {
		if _, err := batch.SetMaybeSent(hash, rawTx, fakeErr.Error()); err != nil {
			return fmt.Errorf("persist uncertain collect batch broadcast: %w", err)
		}
		if err := chainkitcollecttasks.NewRecord().SetBatchMaybeSent(batchId, hash, fakeErr.Error()); err != nil {
			return fmt.Errorf("persist uncertain collect tasks broadcast: %w", err)
		}
		return nil
	}

	sent, err := batch.SetSent(hash, rawTx)
	if err != nil {
		return fmt.Errorf("persist sent collect batch: %w", err)
	}
	if !sent {
		return errors.New("collect batch state changed after broadcast")
	}
	if err := chainkitcollecttasks.NewRecord().SetBatchSent(batchId, hash); err != nil {
		return fmt.Errorf("persist sent collect tasks: %w", err)
	}
	return nil
}
