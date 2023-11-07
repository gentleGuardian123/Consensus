package DPoS

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shengdoushi/base58"

	"golang.org/x/crypto/ripemd160"
)

const versionByte = 0x6f
const addressChecksumLen = 4
var myAlphabet = base58.BitcoinAlphabet

func Sha256Once(inputByte []byte) []byte {
	sha256Example := sha256.New()
	sha256Example.Write(inputByte)

	return sha256Example.Sum(nil)
}

func Ripemd160Once(inputByte []byte) []byte {
	ripemd160Example := ripemd160.New()
	ripemd160Example.Write(inputByte)

	return ripemd160Example.Sum(nil)
}

func Hash160(inputByte []byte) []byte {

	return Ripemd160Once(Sha256Once(inputByte))
}

func Hash256(inputByte []byte) []byte {

	return Sha256Once(Sha256Once(inputByte))
}

func GetAddress(publicKey string) string {
	bytesOfPublicKey, _ := hex.DecodeString(publicKey)
	fingerprint := Hash160(bytesOfPublicKey)

	concat := append([]byte{versionByte}, fingerprint...)
	sumcheck := Hash256(concat)[:addressChecksumLen]

	unEncodedAddress := append(concat, sumcheck...)

	return base58.Encode(unEncodedAddress, myAlphabet)
}

func toHexInt(n *big.Int) string {
    return fmt.Sprintf("%064x", n)
}

func Compress(publKey ecdsa.PublicKey) string {
	var prefix string
	mod := big.NewInt(0)
	mod.Mod(publKey.Y, big.NewInt(2))
	if mod.Cmp(big.NewInt(0)) == 0 {
		prefix = "02"
	} else {
		prefix = "03"
	}

	return prefix + toHexInt(publKey.X)
}

func GenPublicKey() string {
	// rand.Reader = strings.NewReader(seed)
	privKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		log.Panic(err)	
	}
	publKey := privKey.PublicKey
	return Compress(publKey)
}
