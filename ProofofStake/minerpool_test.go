package PoS

import "testing"

func TestMinerPool(t *testing.T) {
	
	mp := NewMinerPool("test_addresses.json", "test_minersAndBlocks.db")

	mp.AddMiner(1, 4)
	mp.AddMiner(3, 2)

	t.Logf("Total number of miners is: %d", len(mp.Addresses))

	for index, address := range mp.Addresses {
		t.Logf("The %02d-th miner's address is: %s", index+1, address)
	}

	for index, address := range mp.Probability {
		t.Logf("The %02d-th probable miner's address is: %s", index+1, address)
	}

}
