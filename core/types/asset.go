package types

type AssetNamespace string

const (
	AssetNamespaceNative AssetNamespace = "native"
	AssetNamespaceToken  AssetNamespace = "token"
)

// Asset identifies what is being transferred.
// Reference meaning depends on chain + namespace:
// - EVM token: contract address (0x...)
// - TRON token: contract address (T... / hex depending on your convention)
// - Native: "ETH"/"TRX"/"BTC" etc.
type Asset struct {
	Chain     Chain
	Namespace AssetNamespace
	Reference string // contract/mint/native symbol key
	Symbol    string
	Decimals  uint8
}
