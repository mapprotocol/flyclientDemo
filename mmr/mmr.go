package mmr

import (
	"bytes"
	"fmt"
	"github.com/marcopoloprotocol/flyclientDemo/common"
	"github.com/marcopoloprotocol/flyclientDemo/rlp"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"golang.org/x/crypto/sha3"
	"math"
	"math/big"
	"sort"
)

const (
	c      = float64(0.5)
	lambda = uint64(50)
)

func RlpHash(x interface{}) (h common.Hash) {
	hw := sha3.New256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
func equal_hash(h1, h2 common.Hash) bool {
	return bytes.Equal(h1[:], h2[:])
}

type Node struct {
	value      common.Hash
	difficulty *big.Int
	index      uint64 // position in array
}

func (n *Node) getHash() common.Hash {
	return n.value
}
func (n *Node) setHash(h common.Hash) {
	n.value = h
}
func (n *Node) getDifficulty() *big.Int {
	return new(big.Int).Set(n.difficulty)
}
func (n *Node) setDifficulty(td *big.Int) {
	n.difficulty = new(big.Int).Set(td)
}
func (n *Node) setIndex(i uint64) {
	n.index = i
}
func (n *Node) getIndex() uint64 {
	return n.index
}
func (n *Node) clone() *Node {
	return &Node{
		value:      n.value,
		difficulty: new(big.Int).Set(n.difficulty),
		index:      n.index,
	}
}
func (n *Node) hasChildren(m *Mmr) bool {
	elem_node_number, curr_root_node_number, aggr_node_number := n.index, m.getRootNode().getIndex(), uint64(0)
	for {
		if curr_root_node_number > 2 {
			leaf_number := node_to_leaf_number(curr_root_node_number)
			left_tree_leaf_number := get_left_leaf_number(leaf_number)
			left_tree_node_number := leaf_to_node_number(left_tree_leaf_number)
			if (aggr_node_number + curr_root_node_number) == (elem_node_number + 1) {
				return true
			}

			if elem_node_number < (aggr_node_number + left_tree_node_number) {
				// branch left
				curr_root_node_number = left_tree_node_number
			} else {
				// branch right
				curr_root_node_number = curr_root_node_number - left_tree_node_number - 1
				aggr_node_number += left_tree_node_number
			}
		} else {
			break
		}
	}
	return false
}
func (n *Node) getChildren(m *Mmr) (*Node, *Node) {
	elem_node_number, curr_root_node_number, aggr_node_number := n.index, m.getRootNode().getIndex(), uint64(0)

	for {
		if curr_root_node_number > 2 {
			leaf_number := node_to_leaf_number(curr_root_node_number)
			left_tree_leaf_number := get_left_leaf_number(leaf_number)
			left_tree_node_number := leaf_to_node_number(left_tree_leaf_number)

			if (aggr_node_number + curr_root_node_number) == (elem_node_number + 1) {
				leaf_number = node_to_leaf_number(curr_root_node_number)
				left_tree_leaf_number = get_left_leaf_number(leaf_number)
				left_tree_node_number = leaf_to_node_number(left_tree_leaf_number)

				left_node_position := aggr_node_number + left_tree_node_number - 1
				right_node_position := aggr_node_number + curr_root_node_number - 2

				left_elem, right_elem := m.getNode(left_node_position), m.getNode(right_node_position)

				return left_elem, right_elem
			}

			if elem_node_number < (aggr_node_number + left_tree_node_number) {
				// branch left
				curr_root_node_number = left_tree_node_number
			} else {
				// branch right
				curr_root_node_number = curr_root_node_number - left_tree_node_number - 1
				aggr_node_number += left_tree_node_number
			}
		} else {
			break
		}
	}

	panic("This node has no children!")
}

type MerkleProof struct {
	MmrSize uint64
	proofs  []common.Hash
}

func newMerkleProof(MmrSize uint64, proof []common.Hash) *MerkleProof {
	return &MerkleProof{
		MmrSize: MmrSize,
		proofs:  proof,
	}
}
func (m *MerkleProof) verify(root common.Hash, pos uint64, leaf_hash common.Hash) bool {
	peaks := get_peaks(m.MmrSize)
	height := 0
	for _, proof := range m.proofs {
		// verify bagging peaks
		if pos_in_peaks(pos, peaks) {
			if pos == peaks[len(peaks)-1] {
				leaf_hash = merge2(leaf_hash, proof)
			} else {
				leaf_hash = merge2(proof, leaf_hash)
				pos = peaks[len(peaks)-1]
			}
			continue
		}
		// verify merkle path
		pos_height, next_height := pos_height_in_tree(pos), pos_height_in_tree(pos+1)
		if next_height > pos_height {
			// we are in right child
			leaf_hash = merge2(proof, leaf_hash)
			pos += 1
		} else {
			leaf_hash = merge2(leaf_hash, proof)
			pos += parent_offset(height)
		}
		height += 1
	}
	return equal_hash(leaf_hash, root)
}

type proofRes struct {
	h  common.Hash
	td *big.Int
}
type SingleMerkleProof struct {
	MmrSize uint64
	proofs  []*proofRes
	node    *proofRes
	nodePos uint64
}

func newSingleMerkleProof(MmrSize uint64, proof []*proofRes, node *proofRes, pos uint64) *SingleMerkleProof {
	return &SingleMerkleProof{
		MmrSize: MmrSize,
		proofs:  proof,
		node:    node,
		nodePos: pos,
	}
}

func (m *SingleMerkleProof) verify(root common.Hash, rTD *big.Int) bool {
	pos, leafHash, leafTD := m.nodePos, m.node.h, new(big.Int).Set(m.node.td)
	peaks := get_peaks(m.MmrSize)
	height := 0
	for _, proof := range m.proofs {
		// verify bagging peaks
		if pos_in_peaks(pos, peaks) {
			if pos == peaks[len(peaks)-1] {
				leafHash = merge2(leafHash, proof.h)
			} else {
				leafHash = merge2(proof.h, leafHash)
				pos = peaks[len(peaks)-1]
			}
			leafTD = new(big.Int).Add(leafTD, proof.td)
			continue
		}
		// verify merkle path
		posHeight, nextHeight := pos_height_in_tree(pos), pos_height_in_tree(pos+1)
		if nextHeight > posHeight {
			// we are in right child
			leafHash = merge2(proof.h, leafHash)
			pos += 1
		} else {
			leafHash = merge2(leafHash, proof.h)
			pos += parent_offset(height)
		}
		leafTD = new(big.Int).Add(leafTD, proof.td)
		height += 1
	}
	return equal_hash(leafHash, root) && 0 == rTD.Cmp(leafTD)
}

type Proof3 struct {
	rootHash common.Hash
	rootTD   *big.Int
	proofs   []*SingleMerkleProof
}

func newProof3(root common.Hash, td *big.Int, proofs []*SingleMerkleProof) *Proof3 {
	return &Proof3{
		rootHash: root,
		rootTD:   new(big.Int).Set(td),
		proofs:   proofs,
	}
}

type Mmr struct {
	Values  []*Node
	CurSize uint64
	LeafNum uint64
}

func NewMmr() *Mmr {
	return &Mmr{
		Values:  make([]*Node, 0, 0),
		CurSize: 0,
		LeafNum: 0,
	}
}
func (m *Mmr) getNode(pos uint64) *Node {
	if pos > m.CurSize-1 {
		return nil
	}
	return m.Values[pos]
}
func (m *Mmr) getLeafNumber() uint64 {
	return m.LeafNum
}
func (m *Mmr) push(n *Node) *Node {
	height, pos := 0, m.CurSize
	n.index = pos
	m.Values = append(m.Values, n)
	m.LeafNum++
	for {
		if pos_height_in_tree(pos+1) > height {
			pos++
			// calculate pos of left child and right child
			left_pos := pos - parent_offset(height)
			right_pos := left_pos + sibling_offset(height)
			left, right := m.Values[left_pos], m.Values[right_pos]
			parent := merge(left, right)
			// for test
			if parent.getIndex() != pos {
				panic("index not match")
			}
			parent.setIndex(pos)
			m.Values = append(m.Values, parent)
			height++
		} else {
			break
		}
	}
	m.CurSize = pos + 1
	return n
}
func (m *Mmr) getRoot() common.Hash {
	if m.CurSize == 0 {
		return common.Hash{}
	}
	if m.CurSize == 1 {
		return m.Values[0].getHash()
	}
	rootNode := m.bagRHSPeaks2(0, get_peaks(m.CurSize))
	if rootNode != nil {
		return rootNode.getHash()
	} else {
		return common.Hash{}
	}
	// return m.bagRHSPeaks(0, get_peaks(m.CurSize))
}
func (m *Mmr) getRootNode() *Node {
	if m.CurSize == 1 {
		return m.Values[0]
	}
	return m.bagRHSPeaks2(0, get_peaks(m.CurSize))
}
func (m *Mmr) getRootDifficulty() *big.Int {
	if m.CurSize == 0 {
		return nil
	}
	if m.CurSize == 1 {
		return m.Values[0].getDifficulty()
	}
	rootNode := m.bagRHSPeaks2(0, get_peaks(m.CurSize))
	if rootNode != nil {
		return rootNode.getDifficulty()
	}
	return nil
}
func (m *Mmr) bagRHSPeaks(pos uint64, peaks []uint64) common.Hash {
	rhs_peak_hashes := make([]common.Hash, 0, 0)
	for _, v := range peaks {
		if v > pos {
			rhs_peak_hashes = append(rhs_peak_hashes, m.Values[v].getHash())
		}
	}
	for {
		if len(rhs_peak_hashes) <= 1 {
			break
		}
		last := len(rhs_peak_hashes) - 1
		right := rhs_peak_hashes[last]
		rhs_peak_hashes = rhs_peak_hashes[:last]
		last = len(rhs_peak_hashes) - 1
		left := rhs_peak_hashes[last]
		rhs_peak_hashes = rhs_peak_hashes[:last]
		rhs_peak_hashes = append(rhs_peak_hashes, merge2(right, left))
	}
	if len(rhs_peak_hashes) == 1 {
		return rhs_peak_hashes[0]
	} else {
		return common.Hash{}
	}
}
func (m *Mmr) bagRHSPeaks2(pos uint64, peaks []uint64) *Node {
	rhsPeakNodes := make([]*Node, 0, 0)
	for _, v := range peaks {
		if v > pos {
			rhsPeakNodes = append(rhsPeakNodes, m.Values[v])
		}
	}
	for {
		if len(rhsPeakNodes) <= 1 {
			break
		}
		last := len(rhsPeakNodes) - 1
		right := rhsPeakNodes[last]
		rhsPeakNodes = rhsPeakNodes[:last]
		last = len(rhsPeakNodes) - 1
		left := rhsPeakNodes[last]
		rhsPeakNodes = rhsPeakNodes[:last]
		parent := merge(right, left)
		parent.setIndex(right.getIndex() + 1)
		rhsPeakNodes = append(rhsPeakNodes, parent)
	}
	if len(rhsPeakNodes) == 1 {
		return rhsPeakNodes[0]
	}
	return nil
}
func (m *Mmr) getChildByAggrWeightDisc(weight *big.Int) uint64 {
	aggr_weight, aggr_node_number, curr_tree_number := big.NewInt(0), uint64(0), m.LeafNum
	for {
		if curr_tree_number > 1 {
			left_tree_number := curr_tree_number / 2
			if !IsPowerOfTwo(curr_tree_number) {
				left_tree_number = NextPowerOfTwo(curr_tree_number) / 2
			}
			n := m.getNode(GetNodeFromLeaf(aggr_node_number+left_tree_number) - 1)
			if n == nil {
				panic("wrong pos1")
			}
			left_tree_difficulty := n.getDifficulty()
			if weight.Cmp(new(big.Int).Add(aggr_weight, left_tree_difficulty)) >= 0 {
				// branch right
				aggr_node_number += left_tree_number
				left_root_node_number := GetNodeFromLeaf(aggr_node_number) - 1
				n1 := m.getNode(left_root_node_number)
				if n1 == nil {
					panic("wrong pos2")
				}
				aggr_weight = new(big.Int).Add(aggr_weight, n1.getDifficulty())
				curr_tree_number = curr_tree_number - left_tree_number
			} else {
				// branch left
				curr_tree_number = left_tree_number
			}
		} else {
			break
		}
	}
	return aggr_node_number
}
func (m *Mmr) getChildByAggrWeight(weight float64) uint64 {
	root_weight := m.getRootDifficulty()
	v1, _ := new(big.Float).Mul(new(big.Float).SetInt(root_weight), big.NewFloat(weight)).Int64()
	weight_disc := big.NewInt(v1)
	return m.getChildByAggrWeightDisc(weight_disc)
}

///////////////////////////////////////////////////////////////////////////////////////
type VerifyElem struct {
	res         *proofRes
	index       uint64
	leaf_number uint64
}

type ProofElem struct {
	cat     uint8 // 0--root,1--node,2 --child
	res     *proofRes
	right   bool
	leafNum uint64
}
type ProofInfo struct {
	root_hash       common.Hash
	root_difficulty *big.Int
	leaf_number     uint64
	elems           []*ProofElem
}
type ProofElems []*ProofElem

func (p ProofElems) pop_back() *ProofElem {
	if len(p) <= 0 {
		return nil
	}
	index := len(p) - 1
	last := p[index]
	p = append(p[:index], p[index+1:]...)
	return last
}
func (p ProofElems) pop_front() *ProofElem {
	if len(p) <= 0 {
		return nil
	}
	index := 0
	last := p[index]
	p = append(p[:index], p[index+1:]...)
	return last
}
func (p ProofElems) is_empty() bool {
	return len(p) == 0
}

type VerifyElems []*VerifyElem

func (v VerifyElems) pop_back() *VerifyElem {
	if len(v) <= 0 {
		return nil
	}
	index := len(v) - 1
	last := v[index]
	v = append(v[:index], v[index+1:]...)
	return last
}
func (v VerifyElems) is_empty() bool {
	return len(v) == 0
}

type ProofBlock struct {
	number      uint64
	aggr_weight float64
}

func (p *ProofBlock) equal(oth *ProofBlock) bool {
	if oth == nil || p == nil {
		return false
	}
	return p.number == oth.number
}

type ProofBlocks []*ProofBlock

func (p ProofBlocks) pop() *ProofBlock {
	if len(p) <= 0 {
		return nil
	}
	index := len(p) - 1
	last := p[index]
	p = append(p[:index], p[index+1:]...)
	return last
}
func (a ProofBlocks) Len() int           { return len(a) }
func (a ProofBlocks) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ProofBlocks) Less(i, j int) bool { return a[i].number < a[j].number }

func get_root(nodes []*VerifyElem) (common.Hash, *big.Int) {
	tmp := []*VerifyElem{}
	for _, v := range nodes {
		tmp = append(tmp, v)
	}
	tmp_nodes := VerifyElems(tmp)
	for {
		if len(tmp) > 1 {
			node2 := tmp_nodes.pop_back()
			node1 := tmp_nodes.pop_back()
			hash := merge2(node1.res.h, node2.res.h)
			tmp_nodes = append(tmp_nodes, &VerifyElem{
				res: &proofRes{
					h:  hash,
					td: new(big.Int).Add(node1.res.td, node2.res.td),
				},
				index:       math.MaxUint64, // uint64(-1) .. none
				leaf_number: math.MaxUint64, // uint64(-1) .. none
			})
		} else {
			break
		}
	}
	if len(tmp_nodes) >= 1 {
		return tmp_nodes[0].res.h, tmp_nodes[0].res.td
	}
	return common.Hash{}, nil
}

///////////////////////////////////////////////////////////////////////////////////////

func generate_proof_recursive(currentNode *Node, blocks []uint64, proofs []*ProofElem,
	max_left_tree_leaf_number uint64, startDepth int, leaf_number_sub_tree uint64, space uint64,
	m *Mmr) []*ProofElem {
	if !currentNode.hasChildren(m) {
		proofs = append(proofs, &ProofElem{
			cat:     2,
			right:   false,
			leafNum: 0,
			res: &proofRes{
				h:  currentNode.getHash(),
				td: currentNode.getDifficulty(),
			},
		})
		return proofs
	}
	left_node, right_node := currentNode.getChildren(m)
	pos := binary_search(blocks, max_left_tree_leaf_number)
	left, right := splitAt(blocks, pos)
	next_left_leaf_number_subtree := get_left_leaf_number(leaf_number_sub_tree)
	if len(left) != 0 {
		depth := get_depth(next_left_leaf_number_subtree)
		diff := uint64(0)
		if depth >= 1 {
			diff = uint64(math.Pow(float64(2), float64(depth-1)))
		}
		proofs = generate_proof_recursive(left_node, left, proofs,
			max_left_tree_leaf_number-diff,
			startDepth, next_left_leaf_number_subtree,
			space+1, m)
	} else {
		proofs = append(proofs, &ProofElem{
			cat:     1,
			right:   false,
			leafNum: 0,
			res: &proofRes{
				h:  left_node.getHash(),
				td: left_node.getDifficulty(),
			},
		})
	}
	if len(right) != 0 {
		depth := get_depth(leaf_number_sub_tree - next_left_leaf_number_subtree)
		diff := uint64(0)
		if depth >= 1 {
			diff = uint64(math.Pow(float64(2), float64(depth-1)))
		}
		proofs = generate_proof_recursive(right_node, right, proofs,
			max_left_tree_leaf_number+diff, startDepth,
			leaf_number_sub_tree-next_left_leaf_number_subtree,
			space+1, m)
	} else {
		proofs = append(proofs, &ProofElem{
			cat:     1,
			right:   true,
			leafNum: 0,
			res: &proofRes{
				h:  right_node.getHash(),
				td: right_node.getDifficulty(),
			},
		})
	}
	return proofs
}

func (m *Mmr) genProof0(right_difficulty *big.Int, blocks []uint64) *ProofInfo {
	blocks = SortAndRemoveRepeatForBlocks(blocks)
	proofs, rootNode, depth := []*ProofElem{}, m.getRootNode(), get_depth(m.getLeafNumber())
	max_leaf_num := uint64(math.Pow(float64(2), float64(depth-1)))
	proofs = generate_proof_recursive(rootNode, blocks, proofs, max_leaf_num, depth,
		m.getLeafNumber(), 0, m)

	proofs = append(proofs, &ProofElem{
		cat:     0,
		right:   false,
		leafNum: m.getLeafNumber(),
		res: &proofRes{
			h:  rootNode.getHash(),
			td: rootNode.getDifficulty(),
		},
	})
	return &ProofInfo{
		root_hash:       m.getRoot(),
		root_difficulty: m.getRootDifficulty(),
		leaf_number:     m.getLeafNumber(),
		elems:           proofs,
	}
}

func (m *Mmr) CreateNewProof(right_difficulty *big.Int) (*ProofInfo, []uint64, []uint64) {
	root_hash := m.getRoot()
	r1, _ := new(big.Float).SetInt(right_difficulty).Float64()
	r2, _ := new(big.Float).SetInt(new(big.Int).Add(m.getRootDifficulty(), right_difficulty)).Float64()
	required_queries := uint64(vd_calculate_m(float64(lambda), c, r1, r2, m.getLeafNumber()) + 1.0)

	weights, blocks := []float64{}, []uint64{}
	for i := 0; i < int(required_queries); i++ {
		h := RlpHash([]interface{}{root_hash, i})
		random := Hash_to_f64(h)
		r3, _ := new(big.Float).SetInt(m.getRootDifficulty()).Float64()
		aggr_weight := cdf(random, vd_calculate_delta(r1, r3))
		weights = append(weights, aggr_weight)
	}
	sort.Float64s(weights)
	for _, v := range weights {
		b := m.getChildByAggrWeight(v)
		blocks = append(blocks, b)
	}
	// Pick up at specific sync point
	// Add extra blocks, which are used for syncing from an already available state
	// 1. block : first block of current 30_000 block interval
	// 2. block : first block of previous 30_000 block interval
	// 3. block : first block of third last 30_000 block interaval
	// 4. block : first block of fourth last 30_000 block interval
	// 5. block : first block of fiftf last 30_000 block interval
	// 6. block : first block of sixth last 30_000 block interval
	// 7. block : first block of seventh last 30_000 block interval
	// 8. block : first block of eighth last 30_000 block interval
	// 9. block : first block of ninth last 30_000 block interval
	// 10. block: first block of tenth last 30_000 block interval
	extra_blocks, current_block := []uint64{}, ((m.getLeafNumber()-1)/30000)*30000
	added := 0
	for {
		if current_block > 30000 && added < 10 {
			blocks = append(blocks, current_block)
			extra_blocks = append(extra_blocks, current_block)
			current_block -= 30000
			added += 1
		} else {
			break
		}
	}

	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i] < blocks[j]
	})
	return m.genProof0(right_difficulty, blocks), blocks, extra_blocks
}

func (p *ProofInfo) verifyProof0(blocks []*ProofBlock) bool {
	blocks = SortAndRemoveRepeatForProofBlocks(blocks)
	blocks = reverseForProofBlocks(blocks)
	proof_blocks := ProofBlocks(blocks)

	proofs := ProofElems(p.elems)
	root_elem := proofs.pop_back()
	if root_elem == nil || root_elem.cat != 0 {
		return false
	}
	if len(proofs) == 1 {
		if it := proofs.pop_back(); it != nil {
			if it.cat == 2 {
				return equal_hash(it.res.h, root_elem.res.h)
			}
		}
		return false
	}
	nodes := VerifyElems([]*VerifyElem{})
	for {
		if !proofs.is_empty() {
			proof_elem := proofs.pop_front()
			if proof_elem.cat == 2 {
				proof_block := proof_blocks.pop()
				number := proof_block.number

				if !nodes.is_empty() {
					//TODO: Verification of previous Mmr should happen here
					//weil in einem Ethereum block header kein Mmr hash vorhanden ist, kann man
					//dies nicht überprüfen, wenn doch irgendwann vorhanden, dann einfach
					//'block_header.Mmr == old_root_hash' überprüfen
					_, left_difficulty := get_root(nodes)
					left, middle := new(big.Float).SetInt(left_difficulty), new(big.Float).Mul(new(big.Float).SetInt(root_elem.res.td), big.NewFloat(proof_block.aggr_weight))
					right := new(big.Float).Add(new(big.Float).SetInt(left_difficulty), new(big.Float).SetInt(proof_elem.res.td))
					if left.Cmp(middle) > 0 || right.Cmp(middle) <= 0 {
						// "aggregated difficulty is not correct, should coincide with: {} <= {} < {}",left, middle, right
						return false
					}
				}
				if number%2 == 0 && number != (root_elem.leafNum-1) {
					right_node := proofs.pop_front()
					right_node_hash, right_node_diff := right_node.res.h, new(big.Int).Set(right_node.res.td)
					if right_node.cat == 2 || right_node.cat == 1 {
						if right_node.cat == 2 {
							proof_blocks.pop()
						}
					} else {
						// Expected ???
						return false
					}
					hash := merge2(proof_elem.res.h, right_node_hash)
					nodes = append(nodes, &VerifyElem{
						res: &proofRes{
							h:  hash,
							td: new(big.Int).Add(proof_elem.res.td, right_node_diff),
						},
						index:       number / 2,
						leaf_number: root_elem.leafNum / 2,
					})
				} else {
					res0 := nodes.pop_back()
					hash := merge2(res0.res.h, proof_elem.res.h)
					nodes = append(nodes, &VerifyElem{
						res: &proofRes{
							h:  hash,
							td: new(big.Int).Add(proof_elem.res.td, res0.res.td),
						},
						index:       number / 2,
						leaf_number: root_elem.leafNum / 2,
					})
				}
			} else if proof_elem.cat == 1 {
				if proof_elem.right {
					left_node := nodes.pop_back()
					hash := merge2(left_node.res.h, proof_elem.res.h)
					nodes = append(nodes, &VerifyElem{
						res: &proofRes{
							h:  hash,
							td: new(big.Int).Add(left_node.res.td, proof_elem.res.td),
						},
						index:       left_node.index / 2,
						leaf_number: left_node.leaf_number / 2,
					})
				} else {
					nodes = append(nodes, &VerifyElem{
						res:         proof_elem.res,
						index:       math.MaxUint64, // UINT64(-1)
						leaf_number: math.MaxUint64, // UINT64(-1)
					})
				}
			} else if proof_elem.cat == 0 {
				// do nothing
			} else {
				panic("invalid cat...")
			}
			for {
				if len(nodes) > 1 {
					node2 := nodes.pop_back()
					node1 := nodes.pop_back()
					if node2.index%2 != 1 && !proofs.is_empty() {
						nodes = append(nodes, node1)
						nodes = append(nodes, node2)
						break
					}
					hash := merge2(node1.res.h, node2.res.h)
					nodes = append(nodes, &VerifyElem{
						res: &proofRes{
							h:  hash,
							td: new(big.Int).Add(node1.res.td, node2.res.td),
						},
						index:       node2.index / 2,
						leaf_number: node2.leaf_number / 2,
					})
				} else {
					break
				}
			}
		} else {
			break
		}
	}

	res0 := nodes.pop_back()
	if res0 != nil {
		return equal_hash(root_elem.res.h, res0.res.h) && root_elem.res.td.Cmp(res0.res.td) == 0
	}
	return false
}
func verify_required_blocks(blocks []uint64, root_hash common.Hash, root_difficulty, right_difficulty *big.Int, root_leaf_number uint64) ([]*ProofBlock, error) {

	r1, _ := new(big.Float).SetInt(right_difficulty).Float64()
	r2, _ := new(big.Float).SetInt(new(big.Int).Add(root_difficulty, right_difficulty)).Float64()
	required_queries := uint64(vd_calculate_m(float64(lambda), c, r1, r2, root_leaf_number) + 1.0)
	extra_blocks, current_block := []uint64{}, ((root_leaf_number-1)/30000)*30000
	added := 0
	for {
		if current_block > 30000 && added < 10 {
			extra_blocks = append(extra_blocks, current_block)
			current_block -= 30000
			added += 1
		} else {
			break
		}
	}

	// required queries can contain the same block number multiple times
	// TODO: maybe multiple blocks can be pruned away?
	if required_queries != uint64(len(blocks)-len(extra_blocks)) {
		return nil, errors.New(fmt.Sprintf("false number of blocks provided: required: %v, got: %v", required_queries, len(blocks)))
	}
	weights := []float64{}
	for i := 0; i < int(required_queries); i++ {
		h := RlpHash([]interface{}{root_hash, i})
		random := Hash_to_f64(h)
		r3, _ := new(big.Float).SetInt(root_difficulty).Float64()
		aggr_weight := cdf(random, vd_calculate_delta(r1, r3))
		weights = append(weights, aggr_weight)
	}
	sort.Float64s(weights)
	proof_blocks, weight_pos := []*ProofBlock{}, 0

	for _, v := range blocks {
		aggr_weight := weights[weight_pos]
		if len(extra_blocks) > 0 {
			index := len(extra_blocks) - 1
			curr_extra_block := extra_blocks[index]
			if v == curr_extra_block {
				extra_blocks = append(extra_blocks[:index], extra_blocks[index+1:]...)
				aggr_weight = 0 // 0--none
			} else {
				weight_pos++
			}
		} else {
			weight_pos++
		}
		proof_blocks = append(proof_blocks, &ProofBlock{
			number:      v,
			aggr_weight: aggr_weight,
		})
	}
	return proof_blocks, nil
}
