package bloom

import (
	"bytes"
	"fmt"
	"github.com/marcopoloprotocol/flyclientDemo/common/hexutil"
	"math/big"
)

type Bloom struct {
	bloom  []byte
	config BloomConfig
}

type BloomConfig struct {
	BloomByteLength uint //
	BloomBits       uint // k个哈希函数
}

func NewBloomConfig(len, bits uint) BloomConfig {
	return BloomConfig{
		BloomBits:       bits,
		BloomByteLength: 1 << (len - 3),
	}
}

func DeriveBloomConfig(count int) BloomConfig {
	l := uint(0)
	count *= 30
	for count != 0 {
		l++
		count >>= 1
	}
	return NewBloomConfig(l, 4)
}

func NewBloom(config BloomConfig) *Bloom {
	return &Bloom{
		bloom:  make([]byte, config.BloomByteLength),
		config: config,
	}
}

func (b *Bloom) GetBloom() []byte {
	return b.bloom
}

func (b *Bloom) String() string {
	return fmt.Sprintf(`%v...%v`, b.bloom[:4], b.bloom[len(b.bloom)-4:])
}

// SetBytes sets the content of b to the given bytes.
func (b *Bloom) SetBytes(d []byte) {
	if len(b.bloom) < len(d) {
		d = d[uint(len(d))-b.config.BloomByteLength:]
	}
	copy(b.bloom[b.config.BloomByteLength-uint(len(d)):], d)
}

// IsEqual tests whether the caller Bloom object is identical
// to the parameter B.
func (b Bloom) IsEqual(B *Bloom) bool {
	return bytes.Equal(b.bloom[:], B.bloom[:])
}

func (b *Bloom) Or(x, y *Bloom) *Bloom {
	for i := range b.bloom {
		b.bloom[i] = x.bloom[i] | y.bloom[i]
	}
	return b
}

func (b Bloom) Hex() string {
	return hexutil.Encode(b.bloom[:])
}

func (b *Bloom) Digest(k []byte) *Bloom {
	hashes := KHash(k[:], b.config.BloomBits)

	for _, hash := range hashes {
		idx := hash & uint32(b.config.BloomByteLength*8-1)

		b.SetAt(idx)
	}

	return b
}

// SetAt sets certain bit to 1 in big-endian
func (b *Bloom) SetAt(idx uint32) {
	op := uint8(1)

	if idx < uint32(b.config.BloomByteLength*8) {
		byteIdx := uint32(b.config.BloomByteLength) - (idx / 8) - 1
		bitIdx := idx % 8

		b.bloom[byteIdx] |= op << bitIdx
	}
}

func (b *Bloom) LookAt(idx uint32) bool {
	op := uint8(1)

	if idx < uint32(b.config.BloomByteLength*8) {
		byteIdx := uint32(b.config.BloomByteLength) - (idx / 8) - 1
		bitIdx := idx % 8

		op <<= bitIdx

		return op == (b.bloom[byteIdx] & op)
	}

	return false
}

// Big converts b to a big integer.
func (b Bloom) Big() *big.Int {
	return new(big.Int).SetBytes(b.bloom[:])
}

func (b Bloom) LookUp(k []byte) bool {
	hashes := KHash(k[:], b.config.BloomBits)

	for _, hash := range hashes {
		idx := hash & uint32(b.config.BloomByteLength*8-1)

		// if not found
		if !b.LookAt(idx) {
			return false
		}
	}

	return true
}

func (b Bloom) Copy() *Bloom {
	cpy := NewBloom(b.config)
	copy(cpy.bloom, b.bloom)
	return cpy
}
