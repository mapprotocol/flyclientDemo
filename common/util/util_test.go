package util

import (
	"fmt"
	"github.com/marcopoloprotocol/flyclientDemo/common/hexutil"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetCurPcIp(t *testing.T) {
	ip := GetCurPcIp("10.200.0")
	fmt.Println(ip)
}

func TestStringifyJsonToBytes(t *testing.T) {
	m := map[string]int{"123": 1}
	var rm map[string]int
	ParseJsonFromBytes(StringifyJsonToBytes(m), &rm)
	assert.Equal(t, m["123"], rm["123"])

	hb, err := hexutil.Decode("0x307830303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030")
	assert.NoError(t, err)
	fmt.Println(hb)
	// 0就是48
	fmt.Println([]byte("0x0000000000000000000000000000000000000000000000"))
	fmt.Println(string(hb))
	fmt.Println(hexutil.Encode(hb))
}

//func TestTxFeeCmp(t *testing.T) {
//	//maxAmount, success := big.NewInt(0).SetString("100000000000000000000", 10)
//	//assert.True(t, success)
//	sumAmount := big.NewInt(3500000000)
//	// x小 -1 ，x大 1
//	fmt.Println(sumAmount.Cmp(big.NewInt(consts.MinAmount)), sumAmount.Cmp(consts.MaxAmount))
//	if sumAmount.Cmp(big.NewInt(consts.MinAmount)) == -1 || sumAmount.Cmp(consts.MaxAmount) == 1 {
//		fmt.Println("false")
//	} else {
//		fmt.Println(true)
//	}
//}

type testRunner struct {
	stopChan chan struct{}
}

func TestStopChanClosed(t *testing.T) {
	runner := &testRunner{}
	assert.True(t, StopChanClosed(runner.stopChan))
	runner.stopChan = make(chan struct{})
	assert.False(t, StopChanClosed(runner.stopChan))
	close(runner.stopChan)
	assert.True(t, StopChanClosed(runner.stopChan))
}

func TestSetTimeout(t *testing.T) {
	stopTimerFunc := SetTimeout(func() {
		fmt.Println("timeout")
	}, time.Second)
	time.Sleep(500 * time.Millisecond)
	stopTimerFunc()

	time.Sleep(time.Second)
}

func TestExecuteFuncWithTimeout(t *testing.T) {
	ExecuteFuncWithTimeout(func() {
		time.Sleep(500 * time.Millisecond)
	}, time.Second)
}

type absHand interface {
	Hit()
}

type hand1 struct {
	X uint
}

func (h *hand1) Hit() {
	fmt.Println(h.X)
}

// 1000000	      1163 ns/op
func BenchmarkInterfaceSliceCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		from := []absHand{&hand1{X: 1}}
		to := make([]*hand1, len(from))
		InterfaceSliceCopy(to, from)
		assert.Equal(b, uint(1), to[0].X)
	}
}

// 1000000	      1004 ns/op
func BenchmarkNormalSliceCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		from := []absHand{&hand1{X: 1}}
		fLen := len(from)
		to := make([]*hand1, fLen)
		for j, f := range from {
			to[j] = f.(*hand1)
		}
		assert.Equal(b, uint(1), to[0].X)
	}
}
