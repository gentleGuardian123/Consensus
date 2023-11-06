package PoS

import "testing"

func TestMain(t *testing.T) {

	mp := NewMinerPool(defaultAddressesJson, defaultDatabase)

	mp.AddMiner(1, 4)
	mp.AddMiner(3, 2)

	mp.ShowAllMiners()

	bc := NewBlockchain(defaultDatabase, mp)

	bc.AddBlock("This is the 1st test data")
	bc.AddBlock("This is the 2nd test data")
	// bc.AddBlock("This is the 3rd test data")
	// bc.AddBlock("This is the 4th test data")
	// bc.AddBlock("This is the 5th test data")
	// bc.AddBlock("This is the 6th test data")
	// bc.AddBlock("This is the 7th test data")
	// bc.AddBlock("This is the 8th test data")
	// bc.AddBlock("This is the 9th test data")

	bc.BlockChainIteration()

}
