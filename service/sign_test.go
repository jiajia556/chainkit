package service

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestSignAndVerifyString(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	service := &ChainService{priKey: key}
	signerAddress := crypto.PubkeyToAddress(key.PublicKey).Hex()

	signature, err := service.SignString("chainkit")
	if err != nil {
		t.Fatalf("sign string: %v", err)
	}
	if len(signature) != recoverableSignatureLength {
		t.Fatalf("signature length = %d, want %d", len(signature), recoverableSignatureLength)
	}

	valid, err := service.VerifyString("chainkit", signature, signerAddress)
	if err != nil {
		t.Fatalf("verify string: %v", err)
	}
	if !valid {
		t.Fatal("signature should be valid")
	}

	valid, err = service.VerifyString("changed", signature, signerAddress)
	if err != nil {
		t.Fatalf("verify changed string: %v", err)
	}
	if valid {
		t.Fatal("signature should not be valid for changed data")
	}
}

func TestSignAndVerifyBytes(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	service := &ChainService{priKey: key}
	signerAddress := crypto.PubkeyToAddress(key.PublicKey).Hex()
	data := []byte{0x00, 0x01, 0xfe, 0xff}

	signature, err := service.SignBytes(data)
	if err != nil {
		t.Fatalf("sign bytes: %v", err)
	}
	valid, err := new(ChainService).VerifyBytes(data, signature, signerAddress)
	if err != nil {
		t.Fatalf("verify bytes: %v", err)
	}
	if !valid {
		t.Fatal("signature should be valid")
	}

	otherKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate other key: %v", err)
	}
	otherAddress := crypto.PubkeyToAddress(otherKey.PublicKey).Hex()
	valid, err = new(ChainService).VerifyBytes(data, signature, otherAddress)
	if err != nil {
		t.Fatalf("verify with other key: %v", err)
	}
	if valid {
		t.Fatal("signature should not be valid for another key")
	}

	stringSignature, err := service.SignString(string(data))
	if err != nil {
		t.Fatalf("sign equivalent string: %v", err)
	}
	if !bytes.Equal(signature, stringSignature) {
		t.Fatal("string and byte signing should use the same encoding")
	}
}

func TestSignAndVerifyPersonalString(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	service := &ChainService{priKey: key}
	signerAddress := crypto.PubkeyToAddress(key.PublicKey).Hex()
	data := "你好, Ethereum"

	signature, err := service.SignPersonalString(data)
	if err != nil {
		t.Fatalf("sign personal string: %v", err)
	}
	expected, err := crypto.Sign(accounts.TextHash([]byte(data)), key)
	if err != nil {
		t.Fatalf("sign expected personal hash: %v", err)
	}
	if !bytes.Equal(signature, expected) {
		t.Fatal("personal signature does not use the EIP-191 text hash")
	}

	valid, err := new(ChainService).VerifyPersonalString(data, signature, signerAddress)
	if err != nil {
		t.Fatalf("verify personal string: %v", err)
	}
	if !valid {
		t.Fatal("personal signature should be valid")
	}

	walletSignature := append([]byte(nil), signature...)
	walletSignature[64] += 27
	walletRecoveryID := walletSignature[64]
	valid, err = new(ChainService).VerifyPersonalString(data, walletSignature, signerAddress)
	if err != nil {
		t.Fatalf("verify personal string with wallet recovery ID: %v", err)
	}
	if !valid {
		t.Fatal("personal signature with recovery ID 27 or 28 should be valid")
	}
	if walletSignature[64] != walletRecoveryID {
		t.Fatal("verification must not mutate the supplied signature")
	}

	valid, err = service.VerifyString(data, signature, signerAddress)
	if err != nil {
		t.Fatalf("verify personal signature as raw signature: %v", err)
	}
	if valid {
		t.Fatal("personal signature should not verify as a raw signature")
	}
}

func TestSignAndVerifyPersonalBytes(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	service := &ChainService{priKey: key}
	signerAddress := crypto.PubkeyToAddress(key.PublicKey).Hex()
	data := []byte{0x00, 0x01, 0xfe, 0xff}

	signature, err := service.SignPersonalBytes(data)
	if err != nil {
		t.Fatalf("sign personal bytes: %v", err)
	}
	valid, err := new(ChainService).VerifyPersonalBytes(data, signature, signerAddress)
	if err != nil {
		t.Fatalf("verify personal bytes: %v", err)
	}
	if !valid {
		t.Fatal("personal byte signature should be valid")
	}

	valid, err = service.VerifyPersonalBytes(append(data, 0x02), signature, signerAddress)
	if err != nil {
		t.Fatalf("verify changed personal bytes: %v", err)
	}
	if valid {
		t.Fatal("personal signature should not be valid for changed data")
	}
}

func TestSignAndVerifyHexSignatures(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	service := &ChainService{priKey: key}
	verifier := new(ChainService)
	signerAddress := crypto.PubkeyToAddress(key.PublicKey).Hex()

	rawSignature, err := service.SignStringHex("chainkit")
	if err != nil {
		t.Fatalf("sign string as hex: %v", err)
	}
	if len(rawSignature) != 2+recoverableSignatureLength*2 || rawSignature[:2] != "0x" {
		t.Fatalf("unexpected hex signature: %q", rawSignature)
	}
	valid, err := verifier.VerifyStringHex("chainkit", rawSignature, signerAddress)
	if err != nil {
		t.Fatalf("verify string hex: %v", err)
	}
	if !valid {
		t.Fatal("hex signature should be valid")
	}

	personalSignature, err := service.SignPersonalBytesHex([]byte("ethereum"))
	if err != nil {
		t.Fatalf("sign personal bytes as hex: %v", err)
	}
	valid, err = verifier.VerifyPersonalBytesHex([]byte("ethereum"), personalSignature, signerAddress)
	if err != nil {
		t.Fatalf("verify personal bytes hex: %v", err)
	}
	if !valid {
		t.Fatal("personal hex signature should be valid")
	}

	if _, err := verifier.VerifyStringHex("chainkit", rawSignature[2:], signerAddress); err == nil {
		t.Fatal("expected missing 0x prefix error")
	}
	if _, err := verifier.VerifyStringHex("chainkit", "0xzz", signerAddress); err == nil {
		t.Fatal("expected invalid hex error")
	}
}

func TestVerifyBytesRejectsMalformedSignature(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	service := &ChainService{priKey: key}
	signerAddress := crypto.PubkeyToAddress(key.PublicKey).Hex()

	if _, err := service.VerifyBytes(nil, make([]byte, 64), signerAddress); err == nil {
		t.Fatal("expected invalid length error")
	}

	signature, err := service.SignBytes(nil)
	if err != nil {
		t.Fatalf("sign empty bytes: %v", err)
	}
	signature[64] = 29
	if _, err := service.VerifyBytes(nil, signature, signerAddress); err == nil {
		t.Fatal("expected invalid recovery ID error")
	}
	if _, err := service.VerifyBytes(nil, make([]byte, recoverableSignatureLength), "not-an-address"); err == nil {
		t.Fatal("expected invalid signer address error")
	}
}

func TestSignAndVerifyRequirePrivateKey(t *testing.T) {
	var nilService *ChainService
	if _, err := nilService.SignBytes(nil); err == nil {
		t.Fatal("expected nil service signing error")
	}
	if _, err := nilService.SignPersonalBytes(nil); err == nil {
		t.Fatal("expected nil service personal signing error")
	}
}
