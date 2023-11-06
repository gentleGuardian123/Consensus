package CMACPoW

import (
	"crypto/aes"
	"encoding/binary"
	"os"
	"testing"

	"github.com/aead/cmac"
)

func TestMain(t *testing.T) {
	var _nonce int64

	for _nonce = 1; _nonce < 100; _nonce ++ {
		nonce := make([]byte, 8)
		binary.LittleEndian.PutUint64(nonce, uint64(_nonce))
		var hash [32]byte
		for i := range hash {
			hash[i] = 113
		}
		block, err := aes.NewCipher(hash[:])
		if err != nil {
			t.Log("Failed to generate the block.")
			os.Exit(1)
		}

		ret, err := cmac.Sum(nonce, block, block.BlockSize())
		if err != nil {
			t.Log("Failed to sum the CMAC of nonce.")
			os.Exit(1)
		}

		t.Logf("The nonce is %x\n", nonce)
		t.Logf("The hash is %x\n", hash)
		t.Logf("The result is %x\n", ret)
		t.Log()
	}

	
}
