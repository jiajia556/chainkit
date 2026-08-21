package deposit

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/models/chainkitchains"
	"github.com/jiajia556/chainkit/models/chainkitdepositeventinbox"
	"github.com/jiajia556/chainkit/models/chainkitdeposittokens"
	"github.com/jiajia556/chainkit/models/chainkittokens"
	"github.com/jiajia556/chainkit/service"
	"github.com/jiajia556/tool-box/log"
	"github.com/shopspring/decimal"
)

const (
	websocketWorkers       = 8
	websocketEventBuffer   = 2048
	websocketRetryMinDelay = time.Second
	websocketRetryMaxDelay = time.Minute
	websocketIdleDelay     = 30 * time.Second
)

var errNoEnabledDepositTokens = errors.New("no enabled deposit tokens")
var errDepositTokenConfigChanged = errors.New("deposit token configuration changed")

type websocketTokenConfig struct {
	addresses         []common.Address
	minDepositAmounts map[common.Address]decimal.Decimal
}

// StartWebSocket runs one reconnecting logs subscription per configured chain.
// It is intentionally independent from Start, which remains the periodic
// getLogs completeness scanner and inbox processor.
func StartWebSocket(ctx context.Context) {
	chains := chainkitchains.NewList()
	if err := chains.FindAll(); err != nil {
		log.Error("failed to load chains for deposit websocket", "error", err)
		return
	}

	var wg sync.WaitGroup
	chains.Foreach(func(_ int, chain *chainkitchains.Record) bool {
		if strings.TrimSpace(chain.Model.WsRpc) == "" {
			log.Debug("deposit websocket is not configured", "chainDbId", chain.Model.Id)
			return true
		}
		wg.Add(1)
		go func(chainDbID uint64) {
			defer wg.Done()
			runChainWebSocket(ctx, chainDbID)
		}(chain.Model.Id)
		return true
	})
	wg.Wait()
}

func runChainWebSocket(ctx context.Context, chainDbID uint64) {
	retryDelay := websocketRetryMinDelay
	for ctx.Err() == nil {
		connectedAt, err := subscribeChainOnce(ctx, chainDbID)
		if ctx.Err() != nil {
			return
		}
		if errors.Is(err, errNoEnabledDepositTokens) {
			retryDelay = websocketRetryMinDelay
			if !waitContext(ctx, websocketIdleDelay) {
				return
			}
			continue
		}
		if errors.Is(err, errDepositTokenConfigChanged) {
			retryDelay = websocketRetryMinDelay
			continue
		}
		if err != nil {
			log.Error("deposit websocket disconnected", "chainDbId", chainDbID, "error", err, "retryIn", retryDelay)
		}
		if time.Since(connectedAt) >= time.Minute {
			retryDelay = websocketRetryMinDelay
		}
		if !waitContext(ctx, retryDelay) {
			return
		}
		if retryDelay < websocketRetryMaxDelay {
			retryDelay *= 2
			if retryDelay > websocketRetryMaxDelay {
				retryDelay = websocketRetryMaxDelay
			}
		}
	}
}

func subscribeChainOnce(ctx context.Context, chainDbID uint64) (time.Time, error) {
	connectedAt := time.Now()
	config, err := loadWebSocketTokenConfig(chainDbID)
	if err != nil {
		return connectedAt, err
	}
	if len(config.addresses) == 0 {
		return connectedAt, errNoEnabledDepositTokens
	}

	chainService, err := service.NewChainService(chainDbID)
	if err != nil {
		return connectedAt, err
	}
	defer chainService.CloseClient()
	if err := chainService.DialWSClient(ctx); err != nil {
		return connectedAt, err
	}

	events := make(chan types.Log, websocketEventBuffer)
	query := ethereum.FilterQuery{
		Addresses: config.addresses,
		Topics:    [][]common.Hash{{transferEventSignature}},
	}
	subscription, err := chainService.GetWSClient().SubscribeFilterLogs(ctx, query, events)
	if err != nil {
		return connectedAt, err
	}
	defer subscription.Unsubscribe()
	connectedAt = time.Now()
	log.Debug("deposit websocket subscribed", "chainDbId", chainDbID, "contractCount", len(config.addresses))

	jobs := make(chan types.Log, websocketEventBuffer)
	var workers sync.WaitGroup
	for i := 0; i < websocketWorkers; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for eventLog := range jobs {
				minAmount, ok := config.minDepositAmounts[eventLog.Address]
				if !ok {
					continue
				}
				eventCtx := WithInboxSource(WithMinDepositAmount(ctx, minAmount), chainkitdepositeventinbox.SourceWebSocket)
				if err := chainService.HandleLog(eventCtx, eventLog.Address.Hex(), service.ModuleDeposit, EnqueueDeposit, eventLog); err != nil {
					log.Error("failed to enqueue deposit websocket event", "chainDbId", chainDbID, "txHash", eventLog.TxHash.Hex(), "logIndex", eventLog.Index, "removed", eventLog.Removed, "error", err)
				}
			}
		}()
	}

	consumeErr := consumeSubscription(ctx, subscription.Err(), events, jobs, chainDbID, config)
	close(jobs)
	workers.Wait()
	return connectedAt, consumeErr
}

func consumeSubscription(ctx context.Context, subscriptionErrors <-chan error, events <-chan types.Log, jobs chan<- types.Log, chainDbID uint64, config *websocketTokenConfig) error {
	configTicker := time.NewTicker(depositTokenCheckInterval)
	defer configTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err, ok := <-subscriptionErrors:
			if !ok || err == nil {
				return errors.New("websocket subscription closed")
			}
			return err
		case eventLog, ok := <-events:
			if !ok {
				return errors.New("websocket event channel closed")
			}
			select {
			case jobs <- eventLog:
			case <-ctx.Done():
				return ctx.Err()
			}
		case <-configTicker.C:
			latestConfig, err := loadWebSocketTokenConfig(chainDbID)
			if err != nil {
				log.Error("failed to refresh deposit websocket token configuration", "chainDbId", chainDbID, "error", err)
				continue
			}
			if len(latestConfig.addresses) == 0 {
				return errNoEnabledDepositTokens
			}
			if !sameWebSocketTokenConfig(config, latestConfig) {
				return errDepositTokenConfigChanged
			}
		}
	}
}

func sameWebSocketTokenConfig(left, right *websocketTokenConfig) bool {
	if left == nil || right == nil || len(left.minDepositAmounts) != len(right.minDepositAmounts) {
		return false
	}
	for address, leftAmount := range left.minDepositAmounts {
		rightAmount, exists := right.minDepositAmounts[address]
		if !exists || !leftAmount.Equal(rightAmount) {
			return false
		}
	}
	return true
}

func loadWebSocketTokenConfig(chainDbID uint64) (*websocketTokenConfig, error) {
	depositTokens := chainkitdeposittokens.NewList().FindAvailableByChainDBID(chainDbID)
	config := &websocketTokenConfig{
		addresses:         make([]common.Address, 0, len(*depositTokens.Records)),
		minDepositAmounts: make(map[common.Address]decimal.Decimal, len(*depositTokens.Records)),
	}
	for _, depositToken := range *depositTokens.Records {
		token := chainkittokens.NewRecord()
		if err := token.Read(depositToken.TokenId); err != nil {
			return nil, err
		}
		if !common.IsHexAddress(token.Model.ContractAddress) || token.Model.ContractAddress == (common.Address{}).Hex() {
			continue
		}
		address := common.HexToAddress(token.Model.ContractAddress)
		if _, exists := config.minDepositAmounts[address]; !exists {
			config.addresses = append(config.addresses, address)
		}
		config.minDepositAmounts[address] = depositToken.MinDepositAmount
	}
	return config, nil
}

func waitContext(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
