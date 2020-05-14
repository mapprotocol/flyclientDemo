package flyclientdemo

import (
	"fmt"
	"github.com/marcopoloprotocol/flyclientDemo/common"
	"github.com/marcopoloprotocol/flyclientDemo/diskdb"
	"github.com/marcopoloprotocol/flyclientDemo/diskdb/memorydb"
	"github.com/marcopoloprotocol/flyclientDemo/mmr"
)

func getDB() diskdb.Database {
	return memorydb.New()
}

type Block struct {
	Nonce   uint64      `json:"nonce"`
	Number  uint64      `json:"height"`
	PreHash common.Hash `json:"parentId"`
	MRoot   common.Hash `json:"m_root"`
}

func NewBlock(num uint64, nonce uint64) *Block {
	return &Block{Nonce: nonce,
		Number: num,
	}
}

func (b Block) Hash() common.Hash {
	return mmr.RlpHash(b)
}

func (b Block) String() string {

	return fmt.Sprintf(`Header(%s):
Height:	        %d
Prehash:        %s
Mmr:            %s
____________________________________________________________
`, b.Hash(), b.Number, b.PreHash, b.MRoot)
}

type BlockChain struct {
	genesis *Block
	blocks  []*Block
	header  *Block
	db      diskdb.Database
	Mmr     *mmr.Mmr
}

var genesisBlock = &Block{
	Nonce:   1,
	Number:  0,
	PreHash: common.Hash{},
	MRoot:   common.Hash{},
}

//func NewBlockChain() (bc *BlockChain) {
//	bc = &BlockChain{
//		header:  genesisBlock,
//		genesis: genesisBlock,
//		blocks:  []*Block{genesisBlock},
//		Mmr:     mmr.NewMmr(),
//		db:      getDB(),
//	}
//	node := &mmr.Node{Value: genesisBlock.Hash(), Content: 0}
//	bc.Mmr.Push(node)
//	ghash := genesisBlock.Hash().Bytes()
//	genc, _ := rlp.EncodeToBytes(genesisBlock)
//	bc.db.Put(ghash, genc)
//	return
//}
//
//func (bc *BlockChain) InsertBlock(b *Block) error {
//	if b.Number == 0 {
//		return errors.New("can not add genesis block")
//	}
//
//	b.PreHash = bc.header.Hash()
//	b.MRoot = bc.Mmr.GetRoot()
//
//	node := &Mmr.Node{Value: b.Hash(), Content: b.Number}
//	bc.Mmr.Push(node)
//
//	//bc.header = bc.blocks[len(bc.blocks)]
//	enc, err := rlp.EncodeToBytes(b)
//	if err != nil {
//		return err
//	}
//	bc.db.Put(b.Hash().Bytes(), enc)
//	bc.blocks = append(bc.blocks, b)
//	bc.header = b
//	return nil
//}
//
//func (bc *BlockChain) GetProof(b *Block, end *Block) *Mmr.MerkleProof {
//	pos := Mmr.GetPosByNumber(b.Number)
//
//	sub := bc.Mmr.SubMmr(Mmr.GetSizeByNumber(end.Number - 1))
//
//	proof := sub.Gen_proof(pos)
//	return proof
//}
