package common

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/marcopoloprotocol/flyclientDemo/common/hexutil"
	"golang.org/x/crypto/sha3"
	"math/big"
	"reflect"
	"strings"
	"time"
)

const (
	HashLength    = 32
	AddressLength = 22
	DiffLength    = 4
	WorkerLength  = 2
	MinerLength   = 2
	AgentLength   = 6
	NonceLength   = WorkerLength + MinerLength + AgentLength
)

var (
	hashT       = reflect.TypeOf(Hash{})
	addressT    = reflect.TypeOf(Address{})
	difficultyT = reflect.TypeOf(Difficulty{})
	blocknonceT = reflect.TypeOf(BlockNonce{})
)

// 地址类型
const (
	AddressTypeNormal = 0x0000
	AddressTypePool   = 0x0001
)

type AddressType int

type Hash [HashLength]byte

func (h *Hash) PutUint32(v uint32) {
	h[HashLength-4] = byte(v >> 24)
	h[HashLength-3] = byte(v >> 16)
	h[HashLength-2] = byte(v >> 8)
	h[HashLength-1] = byte(v)
}

func (h *Hash) Lsh(n int) {
	shift := n / 8
	for i := 0; i < HashLength-shift; i++ {
		h[i] = h[i+shift]
	}
	for ; shift > 0; shift-- {
		h[HashLength-shift] = 0
	}
	for s := n % 8; s > 0; s-- {
		for i := 0; i < HashLength-1; i++ {
			h[i] <<= 1
			if h[i+1]&0x80 == 0x80 {
				h[i] |= 0x01
			}
		}
		h[HashLength-1] <<= 1
	}
}

func (h Hash) String() string {
	return h.Hex()
}

func (h Hash) MarshalText() ([]byte, error) {
	hb := (hexutil.Bytes)(h[:])
	return hb.MarshalText()
}

// UnmarshalText parses a hash in hex syntax. This is use for pure text of json string
func (h *Hash) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Hash", input, h[:])
}

// UnmarshalJSON parses a hash in hex syntax. This is use for json string
func (h *Hash) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(hashT, input, h[:])
}

// Returns an exact copy of the provided bytes
func CopyHash(h *Hash) *Hash {
	copied := Hash{}
	if len(h) == 0 {
		return &copied
	}
	copy(copied[:], h[:])
	return &copied
}

func (h Hash) IsEqual(oh Hash) bool {
	return bytes.Equal(h[:], oh[:])
}
func (h Hash) IsEmpty() bool {
	emptyHash := Hash{}
	return bytes.Equal(h[:], emptyHash[:])
}

// UnprefixedHash allows marshaling a Hash without 0x prefix.
type UnprefixedHash Hash

// UnmarshalText decodes the hash from hex. The 0x prefix is optional.
func (h *UnprefixedHash) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedUnprefixedText("UnprefixedHash", input, h[:])
}

// MarshalText encodes the hash as hex.
func (h UnprefixedHash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

// compare hash h with hash oh, if h bigger than oh return 1, if h equal to oh return 0,if h smaller than oh return -1
func (h Hash) Cmp(oh Hash) int {
	for i := range h {
		if h[i] > oh[i] {
			return 1
		} else if h[i] < oh[i] {
			return -1
		}
	}
	return 0
}

func (h *Hash) Clear() {
	*h = Hash{}
}

// Serialized a struct using rlp encode and hash it
func RlpHashKeccak256(v interface{}) (h Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, v)
	hw.Sum(h[:0])
	return
}

func BytesToHash(b []byte) (result Hash) {
	result.SetBytes(b)
	return
}

// Sets the hash to the value of b. If b is larger than len(h), 'b' will be cropped (from the left).
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

func (h Hash) Hex() string          { return hexutil.Encode(h[:]) }
func (h Hash) HexWithout0x() string { return hexutil.EncodeWithout0x(h[:]) }
func (h Hash) Bytes() []byte        { return h[:] }
func (h Hash) Big() *big.Int        { return new(big.Int).SetBytes(h[:]) }

func HexToHash(s string) Hash   { return BytesToHash(FromHex(s)) }
func BigToHash(b *big.Int) Hash { return BytesToHash(b.Bytes()) }

// hashlevel compute the level of the hash compared to the maxhash,
// when call this function we assume maxhash is bigger or equal to the hash.
func (h Hash) HashLevel(maxHash Hash) int {
	if h.Cmp(maxHash) > 0 {
		return 0
	}

	pos := 1
	cnt := uint(0)
	for ; (cnt < 32*8) && ((maxHash[cnt/8] & (byte(0x80) >> (cnt % 8))) == 0); cnt++ {
	}
	for ; (cnt < 32*8) && ((h[cnt/8] & (byte(0x80) >> (cnt % 8))) == 0); cnt++ {
		pos++
	}
	return pos
}

type Address [AddressLength]byte

func (h Hash) ValidHashForDifficulty(difficulty Difficulty) bool {
	//log.Debug("h ValidHashForDifficulty", "hash", h.Hex(), "diff", difficulty.Hex())
	result := h.Cmp(difficulty.DiffToTarget())
	if result <= 0 {
		return true
	} else {
		return false
	}
	//prefix := strings.Repeat("0", difficulty)
	//return strings.HasPrefix(h.HexWithout0x(), prefix)
}

func (addr Address) String() string {
	return addr.Hex()
}

// MarshalText returns the hex representation of address.
func (addr Address) MarshalText() ([]byte, error) {
	return hexutil.Bytes(addr[:]).MarshalText()
}

// UnmarshalText parses a address in hex syntax.
func (addr *Address) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Address", input, addr[:])
}

// UnmarshalJSON parses a address in hex syntax.
func (addr *Address) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(addressT, input, addr[:])
}

//todo address type is different from tx type, what does tx type really depend?
func (addr Address) Type() AddressType {
	return AddressType(binary.BigEndian.Uint16(addr[:2]))
}

// get description of the address type
func (addr Address) GetAddressTypeStr() string {
	switch addr.Type() {
	case AddressTypeNormal:
		return "Normal Address"
	case AddressTypePool:
		return "Pool Address"
	}
	return "UnKnown"
}

func (addr Address) IsEmpty() bool {
	emptyAddress := Address{}
	if bytes.Equal(addr[:], emptyAddress[:]) {
		return true
	}
	return false
}

func (addr Address) IsEqual(oaddr Address) bool {
	return bytes.Equal(addr[:], oaddr[:])
}

// IsHexAddress verifies whether a string can represent a valid hex-encoded
// address or not.
func IsHexAddress(s string) bool {
	if hasHexPrefix(s) {
		s = s[2:]
	}
	return len(s) == 2*AddressLength && isHex(s)
}

func (addr *Address) Clear() {
	*addr = Address{}
}

func BytesToAddress(b []byte) (result Address) {
	result.SetBytes(b)
	return
}

// Sets the address to the value of b. If b is larger than len(a) it will panic
func (addr *Address) SetBytes(b []byte) {
	if len(b) > len(addr) {
		b = b[len(b)-AddressLength:]
	}
	copy(addr[AddressLength-len(b):], b)
}

// Get the string representation of the underlying address
func (addr Address) Bytes() []byte { return addr[:] }
func (addr Address) Big() *big.Int { return new(big.Int).SetBytes(addr[:]) }
func (addr Address) Hash() Hash    { return BytesToHash(addr[:]) }

func BigToAddress(b *big.Int) Address { return BytesToAddress(b.Bytes()) }
func HexToAddress(s string) Address   { return BytesToAddress(FromHex(s)) }

func (addr Address) Hex() string {
	unchecksummed := hex.EncodeToString(addr[:])
	sha := sha3.NewLegacyKeccak256()
	sha.Write([]byte(unchecksummed))
	hash := sha.Sum(nil)

	result := []byte(unchecksummed)
	for i := 0; i < len(result); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}
	return "0x" + string(result)
}

// UnprefixedAddress allows marshaling an Address without 0x prefix.
type UnprefixedAddress Address

// UnmarshalText decodes the address from hex. The 0x prefix is optional.
func (a *UnprefixedAddress) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedUnprefixedText("UnprefixedAddress", input, a[:])
}

// MarshalText encodes the address as hex.
func (a UnprefixedAddress) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(a[:])), nil
}

// MixedcaseAddress retains the original string, which may or may not be
// correctly checksummed
type MixedcaseAddress struct {
	addr     Address
	original string
}

// UnmarshalJSON parses MixedcaseAddress
func (ma *MixedcaseAddress) UnmarshalJSON(input []byte) error {
	if err := hexutil.UnmarshalFixedJSON(addressT, input, ma.addr[:]); err != nil {
		return err
	}
	return json.Unmarshal(input, &ma.original)
}

// MarshalJSON marshals the original value
func (ma *MixedcaseAddress) MarshalJSON() ([]byte, error) {
	if strings.HasPrefix(ma.original, "0x") || strings.HasPrefix(ma.original, "0X") {
		return json.Marshal(fmt.Sprintf("0x%s", ma.original[2:]))
	}
	return json.Marshal(fmt.Sprintf("0x%s", ma.original))
}

// NewMixedcaseAddress constructor (mainly for testing)
func NewMixedcaseAddress(addr Address) MixedcaseAddress {
	return MixedcaseAddress{addr: addr, original: addr.Hex()}
}

// NewMixedcaseAddressFromString is mainly meant for unit-testing
func NewMixedcaseAddressFromString(hexaddr string) (*MixedcaseAddress, error) {
	if !IsHexAddress(hexaddr) {
		return nil, fmt.Errorf("Invalid address")
	}
	a := FromHex(hexaddr)
	return &MixedcaseAddress{addr: BytesToAddress(a), original: hexaddr}, nil
}

// Address returns the address
func (ma *MixedcaseAddress) Address() Address {
	return ma.addr
}

// String implements fmt.Stringer
func (ma *MixedcaseAddress) String() string {
	if ma.ValidChecksum() {
		return fmt.Sprintf("%s [chksum ok]", ma.original)
	}
	return fmt.Sprintf("%s [chksum INVALID]", ma.original)
}

// ValidChecksum returns true if the address has valid checksum
func (ma *MixedcaseAddress) ValidChecksum() bool {
	return ma.original == ma.addr.Hex()
}

// Original returns the mixed-case input string
func (ma *MixedcaseAddress) Original() string {
	return ma.original
}

// We use the Bitcoin compact difficulty target here.
// See https://bitcoin.org/en/developer-reference#target-nbits for details
type Difficulty [DiffLength]byte

// MarshalText returns the hex representation of diff.
func (diff Difficulty) MarshalText() ([]byte, error) {
	return hexutil.Bytes(diff[:]).MarshalText()
}

// UnmarshalText parses a diff in hex syntax.
func (diff *Difficulty) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Difficulty", input, diff[:])
}

// UnmarshalJSON parses a diff in hex syntax.
func (diff *Difficulty) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(difficultyT, input, diff[:])
}

func (diff Difficulty) IsEqual(d2 Difficulty) bool {
	return bytes.Equal(diff[:], d2[:])
}

func BytesToDiff(b []byte) (result Difficulty) {
	result.SetBytes(b)
	return
}

func (diff *Difficulty) SetBytes(b []byte) {
	if len(b) > len(diff) {
		b = b[len(b)-DiffLength:]
	}
	copy(diff[DiffLength-len(b):], b)
}

func (diff Difficulty) Hex() string          { return hexutil.Encode(diff[:]) }
func (diff Difficulty) HexWithout0x() string { return hexutil.EncodeWithout0x(diff[:]) }
func (diff Difficulty) String() string       { return string(diff[:]) }
func (diff Difficulty) Bytes() []byte        { return diff[:] }

func (diff Difficulty) Big() *big.Int { return diff.DiffToTarget().Big() }

func HexToDiff(s string) Difficulty { return BytesToDiff(FromHex(s)) }

func BigToDiff(b *big.Int) (diff Difficulty) {
	compact := bigToCompact(b)
	binary.BigEndian.PutUint32(diff[:], compact)
	return diff
}

func (diff Difficulty) DiffToTarget() (target Hash) {
	nbits := binary.BigEndian.Uint32(diff[:])
	size := nbits >> 24
	nWord := nbits & 0x007fffff
	if size <= 3 {
		nWord >>= 8 * (3 - size)
		target.PutUint32(nWord)
	} else {
		target.PutUint32(nWord)
		target.Lsh(int(8 * (size - 3)))
	}
	return
}

//Convert a big interger into a compact form
func bigToCompact(n *big.Int) uint32 {
	// No need to do any work if it's zero.
	if n.Sign() == 0 {
		return 0
	}
	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes.  So, shift the number right or left
	// accordingly.  This is equivalent to:
	// mantissa = mantissa / 256^(exponent-3)
	var mantissa uint32
	exponent := uint(len(n.Bytes()))
	if exponent <= 3 {
		mantissa = uint32(n.Bits()[0])
		mantissa <<= 8 * (3 - exponent)
	} else {
		// Use a copy to avoid modifying the caller's original number.
		tn := new(big.Int).Set(n)
		mantissa = uint32(tn.Rsh(tn, 8*(exponent-3)).Bits()[0])
	}

	// When the mantissa already has the sign bit set, the number is too
	// large to fit into the available 23-bits, so divide the number by 256
	// and increment the exponent accordingly.
	if mantissa&0x00800000 != 0 {
		mantissa >>= 8
		exponent++
	}

	// Pack the exponent, sign bit, and mantissa into an unsigned 32-bit
	// int and return it.
	compact := uint32(exponent<<24) | mantissa
	if n.Sign() < 0 {
		compact |= 0x00800000
	}
	return compact
}

// UnprefixedAddress allows marshaling an Address without 0x prefix.
type UnprefixedDifficulty Difficulty

// UnmarshalText decodes the address from hex. The 0x prefix is optional.
func (a *UnprefixedDifficulty) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedUnprefixedText("UnprefixedDifficulty", input, a[:])
}

// MarshalText encodes the address as hex.
func (a UnprefixedDifficulty) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(a[:])), nil
}

type BlockNonce [NonceLength]byte

// MarshalText returns the hex representation of blockNonce.
func (bn BlockNonce) MarshalText() ([]byte, error) {
	return hexutil.Bytes(bn[:]).MarshalText()
}

// UnmarshalText parses a blockNonce in hex syntax.
func (bn *BlockNonce) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("BlockNonce", input, bn[:])
}

// UnmarshalJSON parses a blockNonce in hex syntax.
func (bn *BlockNonce) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(blocknonceT, input, bn[:])
}

//Compare block nonce
func (bn BlockNonce) IsEqual(obn BlockNonce) bool {
	return bytes.Equal(bn[:], obn[:])
}

func (bn BlockNonce) Hex() string          { return hexutil.Encode(bn[:]) }
func (bn BlockNonce) HexWithout0x() string { return hexutil.EncodeWithout0x(bn[:]) }
func (bn BlockNonce) Bytes() []byte        { return bn[:] }
func (bn BlockNonce) Big() *big.Int        { return new(big.Int).SetBytes(bn[:]) }

func BytesToNonce(b []byte) (result BlockNonce) {
	result.SetBytes(b)
	return
}

func (bn *BlockNonce) SetBytes(b []byte) {
	if len(b) > len(bn) {
		b = b[len(b)-NonceLength:]
	}
	copy(bn[NonceLength-len(b):], b)
}

func HexToNonce(s string) BlockNonce   { return BytesToNonce(FromHex(s)) }
func BigToNonce(b *big.Int) BlockNonce { return BytesToNonce(b.Bytes()) }

// Returns an exact copy of the provided bytes
func CopyNonce(n BlockNonce) BlockNonce {
	copied := BlockNonce{}
	if len(n) == 0 {
		return copied
	}
	copy(copied[:], n[:])
	return copied
}

// PrettyAge is a pretty printed version of a time.Duration value that rounds
// the values up to a single most significant unit, days/weeks/years included.
type PrettyAge time.Time

// ageUnits is a list of units the age pretty printing uses.
var ageUnits = []struct {
	Size   time.Duration
	Symbol string
}{
	{12 * 30 * 24 * time.Hour, "y"},
	{30 * 24 * time.Hour, "mo"},
	{7 * 24 * time.Hour, "w"},
	{24 * time.Hour, "d"},
	{time.Hour, "h"},
	{time.Minute, "m"},
	{time.Second, "s"},
}

// String implements the Stringer interface, allowing pretty printing of duration
// values rounded to the most significant time unit.
func (t PrettyAge) String() string {
	// Calculate the time difference and handle the 0 cornercase
	diff := time.Since(time.Time(t))
	if diff < time.Second {
		return "0"
	}
	// Accumulate a precision of 3 components before returning
	result, prec := "", 0

	for _, unit := range ageUnits {
		if diff > unit.Size {
			result = fmt.Sprintf("%s%d%s", result, diff/unit.Size, unit.Symbol)
			diff %= unit.Size

			if prec += 1; prec >= 3 {
				break
			}
		}
	}
	return result
}

// UnprefixedAddress allows marshaling an Address without 0x prefix.
type UnprefixedNonce BlockNonce

// UnmarshalText decodes the address from hex. The 0x prefix is optional.
func (a *UnprefixedNonce) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedUnprefixedText("UnprefixedNonce", input, a[:])
}

// MarshalText encodes the address as hex.
func (a UnprefixedNonce) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(a[:])), nil
}
