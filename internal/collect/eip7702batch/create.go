package eip7702batch

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/models/chainkitcollectbatches"
	"github.com/jiajia556/chainkit/models/chainkitcollectbatchitems"
	"github.com/jiajia556/chainkit/models/chainkitcollecttasks"
	"github.com/jiajia556/tool-box/mysqlx"
)

const maxBatchItems = 100

type CreateRequest struct {
	ChainDbId                uint64
	SponsorMnemonicAddressId uint64
	SponsorAddress           string
	SponsorNonce             uint64
	ExecutorAddress          string
	TxType                   uint8
	Items                    []ItemRequest
}

type ItemRequest struct {
	CollectTaskId         uint64
	UserDepositAddressId  uint64
	AuthorityAddress      string
	AuthorizationRequired bool
	AuthorizationNonce    uint64
	DelegateAddress       string
	TokenAddress          string
	RecipientAddress      string
	PlannedAmount         string
	CallGasLimit          uint64
}

// Create atomically persists a sponsored batch and reserves all source collect
// tasks. The caller should retry from fresh chain state when this returns an error.
func Create(request CreateRequest) (*chainkitcollectbatches.Record, error) {
	prepared, err := prepare(request)
	if err != nil {
		return nil, err
	}

	session := mysqlx.NewTxSession()
	if err := session.Begin(); err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = session.Rollback()
		}
	}()

	batch := chainkitcollectbatches.NewRecord(session)
	batch.Model = prepared.batch
	if err := batch.Create(); err != nil {
		return nil, fmt.Errorf("create collect batch: %w", err)
	}

	for _, itemModel := range prepared.items {
		itemModel.BatchId = batch.Model.Id
		item := chainkitcollectbatchitems.NewRecord(session)
		item.Model = itemModel
		if err := item.Create(); err != nil {
			return nil, fmt.Errorf("create collect batch item for task %d: %w", itemModel.CollectTaskId, err)
		}
	}

	affected, err := chainkitcollecttasks.NewRecord(session).
		BatchClaimForEIP7702(prepared.taskIDs, batch.Model.Id)
	if err != nil {
		return nil, fmt.Errorf("claim collect tasks: %w", err)
	}
	if affected != int64(len(prepared.taskIDs)) {
		return nil, fmt.Errorf(
			"claim collect tasks: affected %d, expected %d",
			affected,
			len(prepared.taskIDs),
		)
	}

	if err := session.Commit(); err != nil {
		return nil, fmt.Errorf("commit collect batch: %w", err)
	}
	committed = true
	return batch, nil
}

type preparedCreate struct {
	batch   *chainkitcollectbatches.ChainCollectBatches
	items   []*chainkitcollectbatchitems.ChainCollectBatchItems
	taskIDs []uint64
}

func prepare(request CreateRequest) (*preparedCreate, error) {
	if request.ChainDbId == 0 {
		return nil, errors.New("chain db id is zero")
	}
	if request.SponsorMnemonicAddressId == 0 {
		return nil, errors.New("sponsor mnemonic address id is zero")
	}
	if !validNonZeroAddress(request.SponsorAddress) {
		return nil, errors.New("invalid sponsor address")
	}
	if !validNonZeroAddress(request.ExecutorAddress) {
		return nil, errors.New("invalid executor address")
	}
	if request.TxType != types.SetCodeTxType && request.TxType != types.DynamicFeeTxType {
		return nil, fmt.Errorf("unsupported sponsored transaction type %d", request.TxType)
	}
	if len(request.Items) == 0 {
		return nil, errors.New("collect batch items are empty")
	}
	if len(request.Items) > maxBatchItems {
		return nil, fmt.Errorf("collect batch has %d items, maximum is %d", len(request.Items), maxBatchItems)
	}

	taskIDs := make([]uint64, 0, len(request.Items))
	items := make([]*chainkitcollectbatchitems.ChainCollectBatchItems, 0, len(request.Items))
	seenTasks := make(map[uint64]struct{}, len(request.Items))
	authorizations := make(map[common.Address]uint64)
	authorizationRequired := make(map[common.Address]bool)
	authorityDepositIDs := make(map[common.Address]uint64)
	var batchDelegate common.Address

	for index, item := range request.Items {
		if item.CollectTaskId == 0 {
			return nil, fmt.Errorf("item %d collect task id is zero", index)
		}
		if _, exists := seenTasks[item.CollectTaskId]; exists {
			return nil, fmt.Errorf("item %d duplicates collect task %d", index, item.CollectTaskId)
		}
		seenTasks[item.CollectTaskId] = struct{}{}
		if item.UserDepositAddressId == 0 {
			return nil, fmt.Errorf("item %d user deposit address id is zero", index)
		}
		if !validNonZeroAddress(item.AuthorityAddress) {
			return nil, fmt.Errorf("item %d has invalid authority address", index)
		}
		if !validNonZeroAddress(item.DelegateAddress) {
			return nil, fmt.Errorf("item %d has invalid delegate address", index)
		}
		delegate := common.HexToAddress(item.DelegateAddress)
		if index == 0 {
			batchDelegate = delegate
		} else if delegate != batchDelegate {
			return nil, fmt.Errorf("item %d uses a different delegate address", index)
		}
		if !common.IsHexAddress(item.TokenAddress) {
			return nil, fmt.Errorf("item %d has invalid token address", index)
		}
		if !validNonZeroAddress(item.RecipientAddress) {
			return nil, fmt.Errorf("item %d has invalid recipient address", index)
		}
		amount, ok := new(big.Int).SetString(item.PlannedAmount, 10)
		if !ok || amount.Sign() <= 0 {
			return nil, fmt.Errorf("item %d has invalid planned amount", index)
		}
		if item.CallGasLimit == 0 {
			return nil, fmt.Errorf("item %d call gas limit is zero", index)
		}

		authority := common.HexToAddress(item.AuthorityAddress)
		if previousID, exists := authorityDepositIDs[authority]; exists && previousID != item.UserDepositAddressId {
			return nil, fmt.Errorf("authority %s has inconsistent deposit address ids", authority.Hex())
		}
		authorityDepositIDs[authority] = item.UserDepositAddressId
		if previousRequired, exists := authorizationRequired[authority]; exists && previousRequired != item.AuthorizationRequired {
			return nil, fmt.Errorf("authority %s has inconsistent authorization requirements", authority.Hex())
		}
		authorizationRequired[authority] = item.AuthorizationRequired
		if item.AuthorizationRequired {
			if previousNonce, exists := authorizations[authority]; exists && previousNonce != item.AuthorizationNonce {
				return nil, fmt.Errorf("authority %s has inconsistent authorization nonces", authority.Hex())
			}
			authorizations[authority] = item.AuthorizationNonce
		}

		taskIDs = append(taskIDs, item.CollectTaskId)
		items = append(items, &chainkitcollectbatchitems.ChainCollectBatchItems{
			CollectTaskId:         item.CollectTaskId,
			ItemIndex:             uint32(index),
			UserDepositAddressId:  item.UserDepositAddressId,
			AuthorityAddress:      common.HexToAddress(item.AuthorityAddress).Hex(),
			AuthorizationRequired: item.AuthorizationRequired,
			AuthorizationNonce:    item.AuthorizationNonce,
			DelegateAddress:       common.HexToAddress(item.DelegateAddress).Hex(),
			TokenAddress:          common.HexToAddress(item.TokenAddress).Hex(),
			RecipientAddress:      common.HexToAddress(item.RecipientAddress).Hex(),
			PlannedAmount:         amount.String(),
			CallGasLimit:          item.CallGasLimit,
			Status:                chainkitcollectbatchitems.StatusWaiting,
		})
	}

	if request.TxType == types.SetCodeTxType && len(authorizations) == 0 {
		return nil, errors.New("type-4 batch requires at least one authorization")
	}
	if request.TxType == types.DynamicFeeTxType && len(authorizations) != 0 {
		return nil, errors.New("type-2 batch cannot contain authorizations")
	}

	return &preparedCreate{
		batch: &chainkitcollectbatches.ChainCollectBatches{
			ChainDbId:                request.ChainDbId,
			SponsorMnemonicAddressId: request.SponsorMnemonicAddressId,
			SponsorAddress:           common.HexToAddress(request.SponsorAddress).Hex(),
			SponsorNonce:             request.SponsorNonce,
			ExecutorAddress:          common.HexToAddress(request.ExecutorAddress).Hex(),
			TxType:                   request.TxType,
			AuthorizationCount:       uint32(len(authorizations)),
			GasLimit:                 0,
			MaxFeePerGas:             "0",
			MaxPriorityFeePerGas:     "0",
			Status:                   chainkitcollectbatches.StatusWaiting,
		},
		items:   items,
		taskIDs: taskIDs,
	}, nil
}

func validNonZeroAddress(address string) bool {
	return common.IsHexAddress(address) && common.HexToAddress(address) != (common.Address{}) && strings.TrimSpace(address) != ""
}
