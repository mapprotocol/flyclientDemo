package mmr

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/marcopoloprotocol/flyclientDemo/common"
)

func TestBytesToHash(t *testing.T) {
	total := 10000
	weights := make([]float64, 0, 0)
	tt := make(map[int]int, 10)
	for i := 0; i < total; i++ {
		h := RlpHash(uint64(i))
		random := Hash_to_f64(h)

		index := int(random * 10)
		tt[index]++

		aggr_weight := cdf(random, 0.000000000000000000000000000000001)
		weights = append(weights, aggr_weight)
	}
	//sort.Float64s(weights)
	fmt.Println(tt)
	res := make(map[int]int, 10)
	for _, weight := range weights {
		index := int(weight * 10)
		res[index]++
	}
	fmt.Println(res)
}

func cdf2(y, lamda float64) float64 {
	return math.Log(1-y) * (-1 / lamda)
}

func Test04(t *testing.T) {
	right_difficulty, root_difficulty := big.NewInt(int64(10000)), big.NewInt(int64(10000000000000))
	lambda, C, leaf_number := uint64(50), float64(float64(50)/100.0), uint64(10)

	aa := [32]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	root_hash := common.BytesToHash(aa[:])
	fmt.Println("root_hash:", root_hash)

	r1, _ := new(big.Float).SetInt(right_difficulty).Float64()
	r2, _ := new(big.Float).SetInt(new(big.Int).Add(root_difficulty, right_difficulty)).Float64()
	fmt.Println("r1:", r1, "r2:", r2)

	required_queries := uint64(vd_calculate_m(float64(lambda), C, r1, r2, leaf_number) + 1.0)

	fmt.Println("required_queries:", required_queries)
	tt := make(map[int]int, 10)
	weights := []float64{}
	for i := 0; i < int(required_queries); i++ {
		h := RlpHash([]interface{}{root_hash, uint64(i)})
		random := Hash_to_f64(h)
		index := int(random * 10)
		tt[index]++
		r3, _ := new(big.Float).SetInt(root_difficulty).Float64()
		aggr_weight := cdf(random, vd_calculate_delta(r1, r3))
		weights = append(weights, aggr_weight)
		//fmt.Println("i:", i, "aggr_weight:", aggr_weight)
	}

	res := make(map[int]int, 10)
	for _, weight := range weights {
		index := int(weight * 10)
		res[index]++
	}
	fmt.Println(tt)
	fmt.Println(res)
	fmt.Println("finish")
}
