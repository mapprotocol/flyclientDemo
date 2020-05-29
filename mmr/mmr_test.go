package mmr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/marcopoloprotocol/flyclientDemo/common"
)

func IntToBytes(n int) []byte {
	data := int64(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

//func run_Mmr(count int, proof_pos uint64) {
//	m := NewMmr()
//	positions := make([]*Node, 0, 0)
//
//	for i := 0; i < count; i++ {
//		positions = append(positions, m.push(&Node{
//			value:      common.BytesToHash(IntToBytes(i)),
//			difficulty: big.NewInt(0),
//		}))
//	}
//	merkle_root := m.getRoot()
//	// proof
//	pos := positions[proof_pos].index
//	// generate proof for proof_elem
//	proof := m.genProof(pos)
//	// verify proof
//	result := proof.verify(merkle_root, pos, positions[proof_pos].getHash())
//	fmt.Println("result:", result)
//}
//func Test01(t *testing.T) {
//	run_Mmr(10000, 50)
//	fmt.Println("finish")
//}

func Test02(t *testing.T) {
	num := uint64(0)
	a := NextPowerOfTwo(num)
	b := float64(100)
	fmt.Println("b:", math.Log(b), "pos_height:", get_depth(6))
	fmt.Println("aa", a, "isPow:", IsPowerOfTwo(num), "GetNodeFromLeaf:", GetNodeFromLeaf(6))
}
func modify_slice(v []int) []int {
	fmt.Println("len(v):", len(v))
	v = append(v, 100)
	fmt.Println("len(v):", len(v))
	return v
}

func Test03(t *testing.T) {
	val := uint64(0x4029000000000000)
	fmt.Println("val:", val, "fval:", ByteToFloat64(Uint64ToBytes(val)))

	aa := [32]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	fmt.Println("aa", Hash_to_f64(common.BytesToHash(aa[:])))
	fmt.Println("finish")
}

func Test05(t *testing.T) {
	mmr := NewMMR()
	for i := 0; i < 11; i++ {
		mmr.Push(&Node{
			value:      BytesToHash(IntToBytes(i)),
			difficulty: big.NewInt(1000),
		})
	}
	right_difficulty := big.NewInt(1000)
	fmt.Println("leaf_number:", mmr.getLeafNumber(), "root_difficulty:", mmr.GetRootDifficulty())
	proof, blocks, eblocks := mmr.CreateNewProof(right_difficulty)
	fmt.Println("blocks_len:", len(blocks), "blocks:", blocks, "eblocks:", len(eblocks))
	fmt.Println("proof:", proof)
	pBlocks, err := VerifyRequiredBlocks(blocks, proof.RootHash, proof.RootDifficulty, right_difficulty, proof.LeafNumber)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	b := proof.VerifyProof(pBlocks)
	fmt.Println("b:", b)
	fmt.Println("finish")
}
