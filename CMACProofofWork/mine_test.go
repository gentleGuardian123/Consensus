package CMACPoW

import "testing"

func MineOnce() (*Blockchain) {
	bc := NewBlockchain("test_blocks.db")
	bc.AddBlock("Testing difficulty Number is 24 bits")	

	return bc
}

func BenchmarkMine(b *testing.B) {
	bc := NewBlockchain("test_blocks.db")	
	bc.BlockChainIteration()

	for i:= 0; i < b.N; i++ {
		MineOnce()
	}
	
	bc.BlockChainIteration()
}