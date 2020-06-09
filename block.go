package flyclientdemo

import (
	"errors"
	"fmt"
	"github.com/marcopoloprotocol/flyclientDemo/common"
	"github.com/marcopoloprotocol/flyclientDemo/diskdb"
	"github.com/marcopoloprotocol/flyclientDemo/diskdb/memorydb"
	"github.com/marcopoloprotocol/flyclientDemo/mmr"
	"github.com/marcopoloprotocol/flyclientDemo/rlp"
	"math/big"
)

var RightDif = big.NewInt(100000)

func getDB() diskdb.Database {
	return memorydb.New()
}

type Block struct {
	Nonce      uint64      `json:"nonce"`
	Number     uint64      `json:"height"`
	PreHash    common.Hash `json:"parentId"`
	Difficulty *big.Int    `json:"difficulty"`
	MRoot      common.Hash `json:"m_root"`
}

func NewBlock(num uint64, nonce uint64, diff *big.Int) *Block {
	return &Block{Nonce: nonce,
		Number:     num,
		Difficulty: diff,
	}
}

func (b Block) Hash() common.Hash {
	return mmr.RlpHash(b)
}

func (b Block) String() string {

	return fmt.Sprintf(`Header(%s):
Height:	        %d
Prehash:        %s
Difficulty      %s
Mmr:            %s
____________________________________________________________
`, b.Hash(), b.Number, b.PreHash, b.Difficulty, b.MRoot)
}

type BlockChain struct {
	genesis   *Block
	blocks    []*Block
	header    *Block
	db        diskdb.Database
	Mmr       *mmr.Mmr
}

var genesisBlock = &Block{
	Nonce:      1,
	Number:     0,
	PreHash:    common.Hash{},
	Difficulty: big.NewInt(0),
	MRoot:      common.Hash{},
}

func NewBlockChain() (bc *BlockChain) {
	bc = &BlockChain{
		header:  genesisBlock,
		genesis: genesisBlock,
		blocks:  []*Block{genesisBlock},
		Mmr:     mmr.NewMMR(),
		db:      getDB(),
	}
	node := mmr.NewNode(genesisBlock.Hash(), big.NewInt(0))
	bc.Mmr.Push(node)
	ghash := genesisBlock.Hash().Bytes()
	genc, _ := rlp.EncodeToBytes(genesisBlock)
	bc.db.Put(ghash, genc)
	return
}

func (bc *BlockChain) InsertBlock(b *Block) error {
	if b.Number == 0 {
		return errors.New("can not add genesis block")
	}

	b.PreHash = bc.header.Hash()

	b.MRoot = bc.Mmr.GetRoot()

	node := mmr.NewNode(b.Hash(), b.Difficulty)
	bc.Mmr.Push(node)

	//bc.header = bc.blocks[len(bc.blocks)]
	enc, err := rlp.EncodeToBytes(b)
	if err != nil {
		return err
	}
	bc.db.Put(b.Hash().Bytes(), enc)
	bc.blocks = append(bc.blocks, b)
	bc.header = b
	return nil
}

func (bc *BlockChain) Len() int {
	return len(bc.blocks)
}

func (bc *BlockChain) GetTailMmr() *mmr.Mmr {
	m:=bc.Mmr.Copy()
	m.Pop()
	return m
}

func (bc *BlockChain) GetProof() *mmr.ProofInfo {
	m:=bc.GetTailMmr()

	res,_,_:=m.CreateNewProof(RightDif)
	return res
}