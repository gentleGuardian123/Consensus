package DPoS

import (
	"math/rand"
	"testing"
)

func TestMain(t *testing.T) {

	mp := NewMinerPool(defaultAddressesJson, defaultDatabase)

	for i := 0; i < 100; i ++ {
		mp.AddMiner(rand.Int63n(20) + int64(10), rand.Int63n(30) + int64(10))
	}

	mp.ShowAllMiners()

	mp.Vote()

	bc := NewBlockchain(mp)

	bc.AddBlock("This is the 1st test data")
	bc.AddBlock("This is the 2nd test data")
	bc.AddBlock("This is the 3rd test data")
	bc.AddBlock("This is the 4th test data")
	bc.AddBlock("This is the 5th test data")
	bc.AddBlock("This is the 6th test data")
	bc.AddBlock("This is the 7th test data")
	bc.AddBlock("This is the 8th test data")
	bc.AddBlock("This is the 9th test data")

	bc.BlockChainIteration()

}
