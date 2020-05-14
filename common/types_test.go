package common

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
)

var (
	//this is hash of rlp encode string "Hello World!"
	hash1 = HexToHash("0x5f961959398c7f059ff492c568a77e43200157514095b7f4a60788d8a2d6013d")
	//this is hash of rlp encode string "Hello World?"
	hash2  = HexToHash("0x1fbec83034bd56b4db7a0264e434ad2cf326c84c8a9c0f63f84da4d5e06f4e76")
	addr1  = HexToAddress("0x000F9328D55ccb3FCe531f199382339f0E576ee840B1")
	addr2  = HexToAddress("0x000F9328D55ccb3FCe531f199382339f0E576ee840B1")
	nonce1 = HexToNonce("0x000000000000000000000000000000000000000000000fff")
)

func TestHash_Func(t *testing.T) {

	testhash1 := hash1
	testhash2 := hash1
	//testhash1 is equal to testhash2
	assert.Equal(t, true, testhash1.IsEqual(testhash2))
	//testhash1 now is clear to empty,check it
	testhash1.Clear()
	if !testhash1.IsEmpty() {
		t.Errorf("Hash clear does not work")
	}
	//testhash1 is not equal to testhash2 now
	if testhash1.IsEqual(testhash2) {
		t.Errorf("Address1 should be empty and not equal to the old one")
	}
}

//
//func TestAddress_Big(t *testing.T) {
//	h := Hash{}
//	b1 := `"0x` + strings.Repeat("1", 64) + `"`
//	b2 := "0x" + strings.Repeat("1", 64)
//	err := h.UnmarshalJSON([]byte(b1))
//	fmt.Println(b1)
//	fmt.Println(err)
//	fmt.Println(h)
//	err = h.UnmarshalText([]byte(b2))
//	fmt.Println(b2)
//	fmt.Println(err)
//	fmt.Println(h)
//}

func TestHashJsonValidation(t *testing.T) {
	var tests = []struct {
		Prefix string
		Size   int
		Error  string
	}{
		{"", 62, "json: cannot unmarshal hex string without 0x prefix into Go value of type common.Hash"},
		{"0x", 66, "hex string has length 66, want 64 for common.Hash"},
		{"0x", 63, "json: cannot unmarshal hex string of odd length into Go value of type common.Hash"},
		{"0x", 0, "hex string has length 0, want 64 for common.Hash"},
		{"0x", 64, ""},
		{"0X", 64, ""},
	}
	for _, test := range tests {
		input := `"` + test.Prefix + strings.Repeat("0", test.Size) + `"`
		var v Hash
		err := json.Unmarshal([]byte(input), &v)
		if err == nil {
			if test.Error != "" {
				t.Errorf("%s: error mismatch: have nil, want %q", input, test.Error)
			}
		} else {
			if err.Error() != test.Error {
				t.Errorf("%s: error mismatch: have %q, want %q", input, err, test.Error)
			}
		}
	}
}

func TestCopyHash(t *testing.T) {
	a := HexToHash("11")
	//make a hard copy of a
	b := CopyHash(&a)
	//check if the value of a and b point to is equal
	assert.Equal(t, a, *b)
	//reset *b to another value
	*b = HexToHash("22")
	//check if the value of a and b point to is not equal
	assert.NotEqual(t, a, *b)
}

func TestHash_Cmp(t *testing.T) {
	assert.Equal(t, 1, hash1.Cmp(hash2))
	assert.Equal(t, 0, hash1.Cmp(hash1))
	assert.Equal(t, -1, hash2.Cmp(hash1))
}

func BenchmarkHash_Cmp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if hash1.Cmp(hash2) != 1 || hash2.Cmp(hash1) != -1 {
			b.Fatal("panic")
		}
	}

}

func TestHashLevel(t *testing.T) {
	var tests = []struct {
		Hash    Hash
		MaxHash Hash
		Level   int
	}{
		{HexToHash("1"), HexToHash("1"), 1},
		{HexToHash("1"), HexToHash("2"), 2},
		{HexToHash("1"), HexToHash("7"), 3},
		{HexToHash("1"), HexToHash("8"), 4},
		{HexToHash("1"), HexToHash("9"), 4},
		{HexToHash("0"), HexToHash("0"), 1},
		{HexToHash("0"), HexToHash("1"), 2},
		{HexToHash("1"), HexToHash("0"), 0},
		{HexToHash("0x1fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"), HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"), 4},
	}

	for _, test := range tests {
		res := test.Hash.HashLevel(test.MaxHash)
		if res != test.Level {
			t.Errorf("the expect level of hash %v to \n maxhash %v is %d,but we got %d", test.Hash, test.MaxHash, test.Level, res)
		}
	}
}

func TestAddress_Func(t *testing.T) {
	testAddr1 := addr1
	testAddr2 := addr1
	if !testAddr1.IsEqual(testAddr2) {
		t.Errorf("Initial address should be equal")
	}
	testAddr1.Clear()
	if !testAddr1.IsEmpty() {
		t.Errorf("Address clear does not work")
	}
	if testAddr1.IsEqual(testAddr2) {
		t.Errorf("Address1 should be empty and not equal to the old one ")
	}
}

func TestIsHexAddress(t *testing.T) {
	tests := []struct {
		str string
		exp bool
	}{
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed1234", true},
		{"5aaeb6053f3e94c9b9a09f33669435e7ef1beaed1234", true},
		{"0X5aaeb6053f3e94c9b9a09f33669435e7ef1beaed1234", true},
		{"0XAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAaaaa", true},
		{"0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAaaaa", true},
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed1", false},
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beae", false},
		{"5aaeb6053f3e94c9b9a09f33669435e7ef1beaed11", false},
		{"0xxaaeb6053f3e94c9b9a09f33669435e7ef1beaed", false},
	}
	for _, test := range tests {
		if result := IsHexAddress(test.str); result != test.exp {
			t.Errorf("IsHexAddress(%s) == %v; expected %v",
				test.str, result, test.exp)
		}
	}
}

func TestAddressUnmarshalJSON(t *testing.T) {
	var tests = []struct {
		Input     string
		ShouldErr bool
		Output    *big.Int
	}{
		{"", true, nil},
		{`""`, true, nil},
		{`"0x"`, true, nil},
		{`"0x00"`, true, nil},
		{`"0xG000000000000000000000000000000000000000"`, true, nil},
		{`"0x00000000000000000000000000000000000000000000"`, false, big.NewInt(0)},
		{`"0x00000000000000000000000000000000000000000010"`, false, big.NewInt(16)},
	}
	for i, test := range tests {
		var v Address
		err := json.Unmarshal([]byte(test.Input), &v)
		if err != nil && !test.ShouldErr {
			t.Errorf("test #%d: unexpected error: %v", i, err)
		}
		if err == nil {
			if test.ShouldErr {
				t.Errorf("test #%d: expected error, got none", i)
			}
			if v.Big().Cmp(test.Output) != 0 {
				t.Errorf("test #%d: address mismatch: have %v, want %v", i, v.Big(), test.Output)
			}
		}
	}
}

func TestAddressHexChecksum(t *testing.T) {
	var tests = []struct {
		Input  string
		Output string
	}{
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed1234", "0x5aaEB6053f3e94c9b9A09F33669435E7eF1BeAED1234"},
		{"0xfb6916095ca1df60bb79ce92ce3ea74c37c5d3591234", "0xfB6916095ca1dF60bB79CE92Ce3EA74C37c5d3591234"},
		{"0xdbf03b407c01e7cd3cbea99509d93f8dddc8c6fb1234", "0xdbF03B407c01e7cd3cbea99509D93F8dDDC8c6FB1234"},
		{"0xd1220a0cf47c7b9be7a2e6ba89f429762e7b9adb1234", "0xd1220A0CF47C7B9BE7A2E6ba89F429762e7b9ADb1234"},
		// Ensure that non-standard length input values are handled correctly
		{"0xa", "0x0000000000000000000000000000000000000000000A"},
		{"0x0a", "0x0000000000000000000000000000000000000000000A"},
		{"0x00a", "0x0000000000000000000000000000000000000000000A"},
		{"0x000000000000000000000000000000000000000a", "0x0000000000000000000000000000000000000000000A"},
	}
	for i, test := range tests {
		output := HexToAddress(test.Input).Hex()
		if output != test.Output {
			t.Errorf("test #%d: failed to match when it should (%s != %s)", i, output, test.Output)
		}
	}
}

func TestMixedcaseAccount_Address(t *testing.T) {

	var res []struct {
		A     MixedcaseAddress
		Valid bool
	}
	if err := json.Unmarshal([]byte(`[
		{"A" : "0xae967917c465db8578ca9024c205720b1a3651A91234", "Valid": false},
		{"A" : "0xaE967917C465DB8578cA9024C205720b1a3651a91234", "Valid": true},
		{"A" : "0XAe967917c465db8578ca9024c205720b1a3651A91234", "Valid": false},
		{"A" : "0x11111111111111111111122222222222233333231234", "Valid": true}
		]`), &res); err != nil {
		t.Fatal(err)
	}

	for _, r := range res {
		if got := r.A.ValidChecksum(); got != r.Valid {
			t.Errorf("Expected checksum %v, got checksum %v, input %v", r.Valid, got, r.A.String())
		}
	}

	//These should throw exceptions:
	var r2 []MixedcaseAddress
	for _, r := range []string{
		`["0x11111111111111111111122222222222233333"]`,     // Too short
		`["0x111111111111111111111222222222222333332"]`,    // Too short
		`["0x11111111111111111111122222222222233333234"]`,  // Too long
		`["0x111111111111111111111222222222222333332344"]`, // Too long
		`["1111111111111111111112222222222223333323"]`,     // Missing 0x
		`["x1111111111111111111112222222222223333323"]`,    // Missing 0
		`["0xG111111111111111111112222222222223333323"]`,   //Non-hex
	} {
		if err := json.Unmarshal([]byte(r), &r2); err == nil {
			t.Errorf("Expected failure, input %v", r)
		}

	}
}

func TestBigToDiff(t *testing.T) {
	diff1 := HexToDiff("0x201fffff")
	maxPowLimit := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 253), big.NewInt(1))
	diff2 := BigToDiff(maxPowLimit)
	assert.Equal(t, diff1, diff2)
}

func TestDiffJsonValidation(t *testing.T) {
	var tests = []struct {
		Prefix string
		Size   int
		Error  string
	}{
		{"", 6, "json: cannot unmarshal hex string without 0x prefix into Go value of type common.Difficulty"},
		{"0x", 10, "hex string has length 10, want 8 for common.Difficulty"},
		{"0x", 7, "json: cannot unmarshal hex string of odd length into Go value of type common.Difficulty"},
		{"0x", 0, "hex string has length 0, want 8 for common.Difficulty"},
		{"0x", 8, ""},
		{"0X", 8, ""},
	}
	for _, test := range tests {
		input := `"` + test.Prefix + strings.Repeat("0", test.Size) + `"`
		var v Difficulty
		err := json.Unmarshal([]byte(input), &v)
		if err == nil {
			if test.Error != "" {
				t.Errorf("%s: error mismatch: have nil, want %q", input, test.Error)
			}
		} else {
			if err.Error() != test.Error {
				t.Errorf("%s: error mismatch: have %q, want %q", input, err, test.Error)
			}
		}
	}
}

func TestDifficulty_DiffToTarget(t *testing.T) {
	tests := []struct {
		diff string
		hash string
		exp  bool
	}{
		{"0x1743eca9", "0x00000000000000000043eca90000000000000000000000000000000000000000", true},
		{"0x172a4e2f", "0x0000000000000000002a4e2f0000000000000000000000000000000000000000", true},
		{"0x1d00ffff", "0x00000000ffff0000000000000000000000000000000000000000000000000000", true},
		{"0x03000001", "0x0000000000000000000000000000000000000000000000000000000000000001", true},
		{"0x20ffffff", "0x7fffff0000000000000000000000000000000000000000000000000000000000", true},
	}

	for _, test := range tests {
		if result := HexToDiff(test.diff).DiffToTarget().IsEqual(HexToHash(test.hash)); result != test.exp {
			t.Errorf("Difficulty (%s) and Hash (%s) IsEqual should be %v, got %v",
				test.diff, test.hash, test.exp, result)
			fmt.Println(HexToDiff(test.diff).DiffToTarget())
		}
	}
}

func BenchmarkDifficulty_DiffToTarget(b *testing.B) {
	testdiff := HexToDiff("0x1743eca9")
	testhash := HexToHash("0x00000000000000000043eca90000000000000000000000000000000000000000")
	for i := 0; i < b.N; i++ {
		if testhash.Cmp(testdiff.DiffToTarget()) != 0 {
			b.Fatal("panic")
		}
	}
}

func TestBlockNonce_Func(t *testing.T) {
	bn := big.NewInt(4095)
	assert.Equal(t, nonce1.Big(), bn)
	assert.Equal(t, BigToNonce(bn), nonce1)
}

func TestNonceJsonValidation(t *testing.T) {
	var tests = []struct {
		Prefix string
		Size   int
		Error  string
	}{
		{"", NonceLength*2 - 2, "json: cannot unmarshal hex string without 0x prefix into Go value of type common.BlockNonce"},
		{"0x", NonceLength*2 + 4, "hex string has length 24, want 20 for common.BlockNonce"},
		{"0x", NonceLength*2 - 1, "json: cannot unmarshal hex string of odd length into Go value of type common.BlockNonce"},
		{"0x", 0, "hex string has length 0, want 20 for common.BlockNonce"},
		{"0x", NonceLength * 2, ""},
		{"0X", NonceLength * 2, ""},
	}
	for _, test := range tests {
		input := `"` + test.Prefix + strings.Repeat("0", test.Size) + `"`
		var v BlockNonce
		err := json.Unmarshal([]byte(input), &v)
		if err == nil {
			if test.Error != "" {
				t.Errorf("%s: error mismatch: have nil, want %q", input, test.Error)
			}
		} else {
			if err.Error() != test.Error {
				t.Errorf("%s: error mismatch: have %q, want %q", input, err, test.Error)
			}
		}
	}
}

//func BytePlusOne(array []byte) (res []byte) {
//	var index = len(array) - 1
//	for index >= 0 {
//		if array[index] < 255 {
//			array[index]++
//			break
//		} else {
//			array[index] = 0
//			index--
//		}
//	}
//	res = array
//	return
//}
//
//func TestRlpHashKeccak256(t *testing.T) {
//	testbyte := [24]byte{}
//	a := BytePlusOne(testbyte[:])
//	for i := 0; i < 65534; i++ {
//		a = BytePlusOne(testbyte[:])
//	}
//	fmt.Println(a)
//	fmt.Println(big.NewInt(1000).Bytes())
//}
