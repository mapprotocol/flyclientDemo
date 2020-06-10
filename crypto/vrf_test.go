package crypto

import (
	"fmt"
	"github.com/marcopoloprotocol/flyclientDemo/common/hexutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

const testPriv1 = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232031"

func TestEvaluate(t *testing.T) {
	key1, err1 := HexToECDSA(testPriv1)
	assert.NoError(t, err1)
	seed := []byte{1, 2}
	vrf, nizk, err := VRFProve(key1, seed)
	fmt.Println(hexutil.Encode(vrf[:]) )
	fmt.Println(hexutil.Encode(nizk[:]) )
	fmt.Println(err)

	res := append(nizk, vrf...)

	fmt.Println(VRFVerify(&key1.PublicKey, seed, res))
}
