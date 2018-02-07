package ether

import (
	"testing"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
	"github.com/ethereum/go-ethereum/common"
)

func TestEthWorker_Call(t *testing.T) {
	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)
	private_key := strings.TrimPrefix(common.BigToHash(key.D).String(), "0x")
	t.Logf("Address: %s", addr.String())
	t.Logf("Key: %s", private_key)
}