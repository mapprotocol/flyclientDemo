package bloom

import (
	"github.com/marcopoloprotocol/flyclientDemo/rlp"
	"io"
)

// The struct(s) end with -RLP are what actually transmitting
// on the networks. The conversion from original bloom to -RLP
// bloom is implemented, they are used in EncodeRLP and DecodeRLP
// functions which implicitly implements RLP encoder.
type BloomRLP struct {
	Bloom  []byte
	Config BloomConfig
}

func (b *Bloom) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, b.BloomRLP())
}

func (b *Bloom) DecodeRLP(s *rlp.Stream) error {
	var bloom BloomRLP
	err := s.Decode(&bloom)
	bloom.CBloom(b)
	return err
}

func (b Bloom) BloomRLP() *BloomRLP {
	res := BloomRLP{
		Bloom:  make([]byte, len(b.bloom)),
		Config: b.config,
	}
	copy(res.Bloom, b.bloom)

	return &res
}

func (b BloomRLP) CBloom(res *Bloom) *Bloom {
	res.bloom = make([]byte, len(b.Bloom))
	copy(res.bloom, b.Bloom)
	res.config = b.Config
	return res
}
