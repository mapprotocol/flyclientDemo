package flyclientdemo

import (
	"fmt"
	"math/big"
	"testing"
)

func TestBlockChain_GetProof(t *testing.T) {
	bc := NewBlockChain()
	length := 40

	for i := 0; i < length; i++ {
		b := NewBlock(uint64(i), 2,big.NewInt(10000))
		bc.InsertBlock(b)
	}

	fmt.Println(bc.GetProof())




	//var wg sync.WaitGroup
	//
	//fmt.Println("waiting for all goroutine ")
	//for n := 0; n < 1000; n++ {
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		target := rand.Uint64() % uint64(length-1)
	//		var pos = Mmr.GetPosByNumber(uint64(target))
	//		p := bc.GetProof(bc.blocks[target], bc.header)
	//		assert.True(t, p.Verify(bc.header.MRoot, pos, bc.blocks[target].Hash()))
	//	}()
	//}
	//wg.Wait()
	//fmt.Println("All goroutines finished!")
}
