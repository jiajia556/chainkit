package service

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const recoverableSignatureLength = 65

// SignString signs the Keccak-256 hash of data with the private key configured
// on the service. The returned signature is encoded as [R || S || V], where V
// is 0 or 1.
func (s *ChainService) SignString(data string) ([]byte, error) {
	return s.SignBytes([]byte(data))
}

// SignBytes signs the Keccak-256 hash of data with the private key configured
// on the service. The returned signature is encoded as [R || S || V], where V
// is 0 or 1.
func (s *ChainService) SignBytes(data []byte) ([]byte, error) {
	if s == nil || s.priKey == nil {
		return nil, errors.New("no private key")
	}

	return crypto.Sign(crypto.Keccak256(data), s.priKey)
}

// SignStringHex is the hex-string form of SignString.
func (s *ChainService) SignStringHex(data string) (string, error) {
	signature, err := s.SignString(data)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(signature), nil
}

// SignBytesHex is the hex-string form of SignBytes.
func (s *ChainService) SignBytesHex(data []byte) (string, error) {
	signature, err := s.SignBytes(data)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(signature), nil
}

// SignPersonalString signs data using the EIP-191 personal-sign convention.
func (s *ChainService) SignPersonalString(data string) ([]byte, error) {
	return s.SignPersonalBytes([]byte(data))
}

// SignPersonalBytes signs data using the EIP-191 personal-sign convention.
// The signed hash is Keccak-256("\x19Ethereum Signed Message:\n" + len(data) + data).
func (s *ChainService) SignPersonalBytes(data []byte) ([]byte, error) {
	if s == nil || s.priKey == nil {
		return nil, errors.New("no private key")
	}

	return crypto.Sign(accounts.TextHash(data), s.priKey)
}

// SignPersonalStringHex is the hex-string form of SignPersonalString.
func (s *ChainService) SignPersonalStringHex(data string) (string, error) {
	signature, err := s.SignPersonalString(data)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(signature), nil
}

// SignPersonalBytesHex is the hex-string form of SignPersonalBytes.
func (s *ChainService) SignPersonalBytesHex(data []byte) (string, error) {
	signature, err := s.SignPersonalBytes(data)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(signature), nil
}

// VerifyString verifies that signature was produced for data by signerAddress.
func (s *ChainService) VerifyString(data string, signature []byte, signerAddress string) (bool, error) {
	return s.VerifyBytes([]byte(data), signature, signerAddress)
}

// VerifyBytes verifies that signature was produced for data by signerAddress.
// signature must use the [R || S || V] encoding returned by SignBytes.
func (s *ChainService) VerifyBytes(data, signature []byte, signerAddress string) (bool, error) {
	return verifyHash(crypto.Keccak256(data), signature, signerAddress)
}

// VerifyStringHex verifies a 0x-prefixed hex signature for string data.
func (s *ChainService) VerifyStringHex(data, signatureHex, signerAddress string) (bool, error) {
	return s.VerifyBytesHex([]byte(data), signatureHex, signerAddress)
}

// VerifyBytesHex verifies a 0x-prefixed hex signature for byte data.
func (s *ChainService) VerifyBytesHex(data []byte, signatureHex, signerAddress string) (bool, error) {
	signature, err := decodeSignatureHex(signatureHex)
	if err != nil {
		return false, err
	}
	return s.VerifyBytes(data, signature, signerAddress)
}

// VerifyPersonalString verifies an EIP-191 personal-sign signature for data
// against signerAddress.
func (s *ChainService) VerifyPersonalString(data string, signature []byte, signerAddress string) (bool, error) {
	return s.VerifyPersonalBytes([]byte(data), signature, signerAddress)
}

// VerifyPersonalBytes verifies an EIP-191 personal-sign signature for data
// against signerAddress.
func (s *ChainService) VerifyPersonalBytes(data, signature []byte, signerAddress string) (bool, error) {
	return verifyHash(accounts.TextHash(data), signature, signerAddress)
}

// VerifyPersonalStringHex verifies a 0x-prefixed EIP-191 signature for string data.
func (s *ChainService) VerifyPersonalStringHex(data, signatureHex, signerAddress string) (bool, error) {
	return s.VerifyPersonalBytesHex([]byte(data), signatureHex, signerAddress)
}

// VerifyPersonalBytesHex verifies a 0x-prefixed EIP-191 signature for byte data.
func (s *ChainService) VerifyPersonalBytesHex(data []byte, signatureHex, signerAddress string) (bool, error) {
	signature, err := decodeSignatureHex(signatureHex)
	if err != nil {
		return false, err
	}
	return s.VerifyPersonalBytes(data, signature, signerAddress)
}

func decodeSignatureHex(signatureHex string) ([]byte, error) {
	signature, err := hexutil.Decode(signatureHex)
	if err != nil {
		return nil, fmt.Errorf("decode signature hex: %w", err)
	}
	return signature, nil
}

func verifyHash(hash, signature []byte, signerAddress string) (bool, error) {
	if !common.IsHexAddress(signerAddress) {
		return false, fmt.Errorf("invalid signer address: %q", signerAddress)
	}
	if len(signature) != recoverableSignatureLength {
		return false, fmt.Errorf("invalid signature length: got %d, want %d", len(signature), recoverableSignatureLength)
	}
	normalizedSignature := signature
	switch signature[64] {
	case 0, 1:
	case 27, 28:
		normalizedSignature = append([]byte(nil), signature...)
		normalizedSignature[64] -= 27
	default:
		return false, fmt.Errorf("invalid signature recovery ID: %d", signature[64])
	}

	publicKey, err := crypto.SigToPub(hash, normalizedSignature)
	if err != nil {
		return false, fmt.Errorf("recover signature public key: %w", err)
	}

	return crypto.PubkeyToAddress(*publicKey) == common.HexToAddress(signerAddress), nil
}
