package eip7702batch

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/models/chainkitcollectbatches"
	"github.com/jiajia556/chainkit/models/chainkitcollectconfig"
	"github.com/jiajia556/chainkit/models/chainkitcollecttasks"
	"github.com/jiajia556/chainkit/models/chainkitmnemonicaddresses"
	"github.com/jiajia556/chainkit/models/chainkittokens"
	"github.com/jiajia556/chainkit/service"
)

const (
	defaultAutomaticBatchItems = 50
	defaultItemCallGasLimit    = 100_000
)

type authorityPlan struct {
	authorizationRequired bool
	authorizationNonce    uint64
}

// RunChain advances existing batches and creates at most one new batch for a
// chain. The caller selects this runner instead of the traditional collector.
func RunChain(
	ctx context.Context,
	config *chainkitcollectconfig.ChainCollectConfig,
	depositPassword string,
	sponsorPassword string,
) error {
	if config == nil || !config.EIP7702Enabled {
		return nil
	}
	if err := validateAutomaticConfig(config); err != nil {
		return err
	}

	for _, batch := range *chainkitcollectbatches.NewList().
		GetBroadcastedList(config.ChainDbId).Records {
		if err := Check(ctx, batch.Id); err != nil {
			return fmt.Errorf("check EIP-7702 batch %d: %w", batch.Id, err)
		}
	}

	sponsor := chainkitmnemonicaddresses.NewRecord()
	if err := sponsor.Read(config.GasProviderMnemonicAddressId); err != nil {
		return fmt.Errorf("read EIP-7702 sponsor: %w", err)
	}
	if !common.IsHexAddress(sponsor.Model.Address) {
		return errors.New("EIP-7702 sponsor has an invalid address")
	}

	inFlight := chainkitcollectbatches.NewRecord().GetInFlightBySponsor(
		config.ChainDbId,
		sponsor.Model.Address,
	)
	if inFlight.Exists() {
		return nil
	}

	waitingBatches := chainkitcollectbatches.NewList().GetWaitingList(config.ChainDbId, 1)
	if !waitingBatches.IsEmpty() {
		return Execute(ctx, (*waitingBatches.Records)[0].Id, depositPassword, sponsorPassword)
	}

	batch, err := buildAutomaticBatch(ctx, config, sponsor.Model.Address)
	if err != nil || batch == nil {
		return err
	}
	return Execute(ctx, batch.Model.Id, depositPassword, sponsorPassword)
}

func buildAutomaticBatch(
	ctx context.Context,
	config *chainkitcollectconfig.ChainCollectConfig,
	sponsorAddress string,
) (*chainkitcollectbatches.Record, error) {
	limit := int(config.EIP7702MaxBatchItems)
	if limit == 0 {
		limit = defaultAutomaticBatchItems
	}
	if limit > maxBatchItems {
		limit = maxBatchItems
	}
	callGasLimit := config.EIP7702CallGasLimit
	if callGasLimit == 0 {
		callGasLimit = defaultItemCallGasLimit
	}

	tasks := chainkitcollecttasks.NewList().GetWaitingUndecidedList(config.ChainDbId, limit)
	if tasks.IsEmpty() {
		return nil, nil
	}

	chainService, err := service.NewChainService(config.ChainDbId)
	if err != nil {
		return nil, err
	}
	defer chainService.CloseClient()

	sponsorNonce, err := chainService.GetClient().PendingNonceAt(ctx, common.HexToAddress(sponsorAddress))
	if err != nil {
		return nil, fmt.Errorf("read EIP-7702 sponsor nonce: %w", err)
	}

	plans := make(map[common.Address]authorityPlan)
	items := make([]ItemRequest, 0, limit)
	hasAuthorization := false
	for _, task := range *tasks.Records {
		authority := common.HexToAddress(task.FromAddress)
		plan, exists := plans[authority]
		if !exists {
			delegation, inspectErr := chainService.DelegationAt(
				ctx,
				task.FromAddress,
				config.EIP7702DelegateAddress,
			)
			if inspectErr != nil {
				return nil, fmt.Errorf("inspect authority %s: %w", task.FromAddress, inspectErr)
			}
			switch delegation.State {
			case service.DelegationNone:
				nonce, nonceErr := chainService.StableAuthorizationNonce(ctx, task.FromAddress)
				if nonceErr != nil {
					return nil, nonceErr
				}
				plan = authorityPlan{authorizationRequired: true, authorizationNonce: nonce}
			case service.DelegationExpected:
				plan = authorityPlan{}
			default:
				// An unexpected code designation is not overwritten automatically.
				// Leave the task unclaimed for operator review.
				continue
			}
			plans[authority] = plan
		}

		tokenAddress := common.Address{}.Hex()
		var balanceErr error
		var balance string
		if task.TokenId == 0 {
			amount, queryErr := chainService.BalanceAt(task.FromAddress)
			balanceErr = queryErr
			balance = amount.String()
		} else {
			token := chainkittokens.NewRecord()
			if readErr := token.Read(task.TokenId); readErr != nil {
				return nil, fmt.Errorf("read token %d: %w", task.TokenId, readErr)
			}
			if token.Model.ChainDbId != config.ChainDbId || !common.IsHexAddress(token.Model.ContractAddress) {
				return nil, fmt.Errorf("token %d does not belong to chain or has invalid address", task.TokenId)
			}
			tokenAddress = common.HexToAddress(token.Model.ContractAddress).Hex()
			amount, queryErr := chainService.BalanceOf(tokenAddress, task.FromAddress)
			balanceErr = queryErr
			balance = amount.String()
		}
		if balanceErr != nil {
			return nil, fmt.Errorf("read collect balance for task %d: %w", task.Id, balanceErr)
		}
		amount, ok := new(big.Int).SetString(balance, 10)
		if !ok || amount.Sign() <= 0 {
			continue
		}

		items = append(items, ItemRequest{
			CollectTaskId:         task.Id,
			UserDepositAddressId:  task.UserDepositAddressId,
			AuthorityAddress:      authority.Hex(),
			AuthorizationRequired: plan.authorizationRequired,
			AuthorizationNonce:    plan.authorizationNonce,
			DelegateAddress:       config.EIP7702DelegateAddress,
			TokenAddress:          tokenAddress,
			RecipientAddress:      task.ToAddress,
			PlannedAmount:         amount.String(),
			CallGasLimit:          callGasLimit,
		})
		if plan.authorizationRequired {
			hasAuthorization = true
		}
	}
	if len(items) == 0 {
		return nil, nil
	}

	txType := uint8(types.DynamicFeeTxType)
	if hasAuthorization {
		txType = types.SetCodeTxType
	}
	return Create(CreateRequest{
		ChainDbId:                config.ChainDbId,
		SponsorMnemonicAddressId: config.GasProviderMnemonicAddressId,
		SponsorAddress:           sponsorAddress,
		SponsorNonce:             sponsorNonce,
		ExecutorAddress:          config.EIP7702ExecutorAddress,
		TxType:                   txType,
		Items:                    items,
	})
}

func validateAutomaticConfig(config *chainkitcollectconfig.ChainCollectConfig) error {
	if config.ChainDbId == 0 {
		return errors.New("EIP-7702 config chain id is zero")
	}
	if config.GasProviderMnemonicAddressId == 0 {
		return errors.New("EIP-7702 sponsor mnemonic address is not configured")
	}
	if !validNonZeroAddress(config.EIP7702DelegateAddress) {
		return errors.New("EIP-7702 delegate address is invalid")
	}
	if !validNonZeroAddress(config.EIP7702ExecutorAddress) {
		return errors.New("EIP-7702 executor address is invalid")
	}
	if config.EIP7702MaxBatchItems > maxBatchItems {
		return fmt.Errorf("EIP-7702 max batch items cannot exceed %d", maxBatchItems)
	}
	return nil
}
