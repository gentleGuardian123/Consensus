package PoS

import (
	"testing"
)

func TestGenPublKey(t *testing.T) {

	publKey := GenPublicKey()

	t.Logf("The public key is %s", publKey)
	t.Log()

	address := GetAddress(publKey)

	t.Logf("The address is %s", address)
}