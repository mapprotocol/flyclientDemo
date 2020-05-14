package util

import (
	"io"
	"math/big"
)

type Token struct {
	Balances map[string]*big.Int
}

type Bmap struct {
	K string
	V *big.Int
}

type TokenRLP struct {
	Balancem []*Bmap
}

func (t *Token) EncodeRLP(w io.Writer) error {
	var bmapSlice []*Bmap

	for k, v := range t.Balances {
		bmapSlice = append(bmapSlice, &Bmap{K: k, V: v})
	}
	return rlp.Encode(w, TokenRLP{
		bmapSlice,
	})
}

//DecodeRLP implements rlp.Decoder
func (t *Token) DecodeRLP(s *rlp.Stream) error {
	var tokenRLP TokenRLP
	if err := s.Decode(&tokenRLP); err != nil {
		return err
	}
	m := make(map[string]*big.Int)
	for _, b := range tokenRLP.Balancem {
		m[b.K] = b.V
	}
	t.Balances = m
	return nil
}
