package flyclientdemo

import (
	"fmt"
	"github.com/marcopoloprotocol/flyclientDemo/mmr"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func TestBlockChain_GetProof(t *testing.T) {
	bc := NewBlockChain()
	length := 500000

	for i := 0; i < length; i++ {
		b := NewBlock(uint64(i), 2,big.NewInt(10000))
		bc.InsertBlock(b)
	}

	start:=time.Now()
	proof:=bc.GetProof()
	fmt.Println(len(proof.Elems))


	fmt.Println("gen proof cost:",time.Now().Sub(start))
	start=time.Now()
	pBlocks, err := mmr.VerifyRequiredBlocks(proof,RightDif)
	assert.NoError(t,err)
	assert.True(t,proof.VerifyProof(pBlocks))


	fmt.Println("verify cost:", time.Now().Sub(start))

	//var wg sync.WaitGroup
	//fmt.Println("waiting for all goroutine ")
	//for n := 0; n < 1000; n++ {
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//
	//	}()
	//}
	//wg.Wait()
	//fmt.Println("All goroutines finished!")
}
