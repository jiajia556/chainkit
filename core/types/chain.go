package types

type ChainType string

const (
	ChainTypeEVM  ChainType = "evm"
	ChainTypeTRON ChainType = "tron"
	ChainTypeBTC  ChainType = "btc" // future
	ChainTypeSOL  ChainType = "sol" // future
)

type Network string

const (
	NetworkMainnet Network = "mainnet"
	NetworkTestnet Network = "testnet"
	NetworkDevnet  Network = "devnet" // e.g. solana
)

type Chain struct {
	Type    ChainType `json:"type" yaml:"type"`
	Network Network   `json:"network" yaml:"network"`
	ChainID uint64    `json:"chain_id" yaml:"chain_id"` // EVM用；TRON/BTC可置0，但建议保留字段统一
	Name    string    `json:"name" yaml:"name"`         // e.g. "eth", "bsc", "tron-shasta"
}
