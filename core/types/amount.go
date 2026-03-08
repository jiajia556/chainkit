package types

import "math/big"

// Amount is represented in base units (integer), plus decimals metadata for formatting.
type Amount struct {
	Base     *big.Int // minimal unit: wei/sun/satoshi/lamport...
	Decimals uint8
}
