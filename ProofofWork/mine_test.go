package PoW

import (
	"fmt"
	"testing"
	"time"
)

func TestMineOnceTime(t *testing.T) {

	bc := NewBlockchain("test_blocks.db")
	start := time.Now()
	for i := 0; i < 9; i ++ {
		bc.AddBlock("Test with difficulty 24(-bit)")
	}
	end := time.Now()
	elapsed := end.Sub(start)
	// Genesis block + 9 blocks
	fmt.Printf("The average time for mining a block with difficulty 16(-bit) is: %.2fs\n", elapsed.Seconds()/10)

} 

