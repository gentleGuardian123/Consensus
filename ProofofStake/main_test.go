package PoS

import (
	"testing"
)

func TestMain(t *testing.T) {

	mp := NewMinerPool(defaultAddressesJson, defaultDatabase)

	mp.AddMiner(1, 4)
	mp.AddMiner(3, 2)

	mp.ShowAllMiners()

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
	bc.AddBlock("This is the 10th test data")
	bc.AddBlock("This is the 11st test data")
	bc.AddBlock("This is the 12nd test data")
	bc.AddBlock("This is the 13rd test data")
	bc.AddBlock("This is the 14th test data")
	bc.AddBlock("This is the 15th test data")
	bc.AddBlock("This is the 16th test data")
	bc.AddBlock("This is the 17th test data")
	bc.AddBlock("This is the 18th test data")
	bc.AddBlock("This is the 19th test data")

	bc.BlockChainIteration()

	bc.ShowBlocksOfMiner(mp.Addresses[0])
	bc.ShowBlocksOfMiner(mp.Addresses[1])

}
