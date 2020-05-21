package mmr

import (
	"bytes"
	"fmt"

	"errors"
	"math"
	"math/big"
	"sort"

	"github.com/marcopoloprotocol/flyclientDemo/common"
	"github.com/marcopoloprotocol/flyclientDemo/rlp"

	"golang.org/x/crypto/sha3"
)

const (
	c      = float64(0.5)
	lambda = uint64(50)
)

func BytesToHash(b []byte) common.Hash {
	var a common.Hash
	a.SetBytes(b)
	return a
}

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

func NewNode(v common.Hash, d *big.Int) *Node {
	return &Node{
		value:      v,
		difficulty: new(big.Int).Set(d),
	}
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
func (n *Node) hasChildren(m *mmr) bool {
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
func (n *Node) getChildren(m *mmr) (*Node, *Node) {
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

/////////////////////////////////////////////////////////////////////////////////
type proofRes struct {
	h  common.Hash
	td *big.Int
}
type VerifyElem struct {
	Res        *proofRes
	Index      uint64
	LeafNumber uint64
}

type ProofElem struct {
	Cat     uint8 // 0--root,1--node,2 --child
	Res     *proofRes
	Right   bool
	LeafNum uint64
}
type ProofInfo struct {
	RootHash       common.Hash
	RootDifficulty *big.Int
	LeafNumber     uint64
	Elems          []*ProofElem
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
	Number     uint64
	AggrWeight float64
}

func (p *ProofBlock) equal(oth *ProofBlock) bool {
	if oth == nil || p == nil {
		return false
	}
	return p.Number == oth.Number
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
func (a ProofBlocks) Less(i, j int) bool { return a[i].Number < a[j].Number }

//////////////////////////////////////////////////////////////////////////////////////

type mmr struct {
	values  []*Node
	curSize uint64
	leafNum uint64
}

func NewMMR() *mmr {
	return &mmr{
		values:  make([]*Node, 0, 0),
		curSize: 0,
		leafNum: 0,
	}
}
func (m *mmr) getNode(pos uint64) *Node {
	if pos > m.curSize-1 {
		return nil
	}
	return m.values[pos]
}
func (m *mmr) getLeafNumber() uint64 {
	return m.leafNum
}
func (m *mmr) push(n *Node) *Node {
	height, pos := 0, m.curSize
	n.index = pos
	m.values = append(m.values, n)
	m.leafNum++
	for {
		if pos_height_in_tree(pos+1) > height {
			pos++
			// calculate pos of left child and right child
			left_pos := pos - parent_offset(height)
			right_pos := left_pos + sibling_offset(height)
			left, right := m.values[left_pos], m.values[right_pos]
			parent := merge(left, right)
			// for test
			if parent.getIndex() != pos {
				panic("index not match")
			}
			parent.setIndex(pos)
			m.values = append(m.values, parent)
			height++
		} else {
			break
		}
	}
	m.curSize = pos + 1
	return n
}
func (m *mmr) getRoot() common.Hash {
	if m.curSize == 0 {
		return common.Hash{0}
	}
	if m.curSize == 1 {
		return m.values[0].getHash()
	}
	rootNode := m.bagRHSPeaks(0, get_peaks(m.curSize))
	if rootNode != nil {
		return rootNode.getHash()
	} else {
		return common.Hash{0}
	}
}
func (m *mmr) getRootNode() *Node {
	if m.curSize == 1 {
		return m.values[0]
	}
	return m.bagRHSPeaks(0, get_peaks(m.curSize))
}
func (m *mmr) getRootDifficulty() *big.Int {
	if m.curSize == 0 {
		return nil
	}
	if m.curSize == 1 {
		return m.values[0].getDifficulty()
	}
	rootNode := m.bagRHSPeaks(0, get_peaks(m.curSize))
	if rootNode != nil {
		return rootNode.getDifficulty()
	}
	return nil
}
func (m *mmr) bagRHSPeaks(pos uint64, peaks []uint64) *Node {
	rhsPeakNodes := make([]*Node, 0, 0)
	for _, v := range peaks {
		if v > pos {
			rhsPeakNodes = append(rhsPeakNodes, m.values[v])
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

func (m *mmr) getChildByAggrWeightDisc(weight *big.Int) uint64 {
	AggrWeight, aggr_node_number, curr_tree_number := big.NewInt(0), uint64(0), m.leafNum
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
			if weight.Cmp(new(big.Int).Add(AggrWeight, left_tree_difficulty)) >= 0 {
				// branch right
				aggr_node_number += left_tree_number
				left_root_node_number := GetNodeFromLeaf(aggr_node_number) - 1
				n1 := m.getNode(left_root_node_number)
				if n1 == nil {
					panic("wrong pos2")
				}
				AggrWeight = new(big.Int).Add(AggrWeight, n1.getDifficulty())
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
func (m *mmr) getChildByAggrWeight(weight float64) uint64 {
	root_weight := m.getRootDifficulty()
	v1, _ := new(big.Float).Mul(new(big.Float).SetInt(root_weight), big.NewFloat(weight)).Int64()
	weight_disc := big.NewInt(v1)
	return m.getChildByAggrWeightDisc(weight_disc)
}

///////////////////////////////////////////////////////////////////////////////////////

func generateProofRecursive(currentNode *Node, blocks []uint64, proofs []*ProofElem,
	max_left_tree_leaf_number uint64, startDepth int, leaf_number_sub_tree uint64, space uint64,
	m *mmr) []*ProofElem {
	if !currentNode.hasChildren(m) {
		proofs = append(proofs, &ProofElem{
			Cat:     2,
			Right:   false,
			LeafNum: 0,
			Res: &proofRes{
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
		proofs = generateProofRecursive(left_node, left, proofs,
			max_left_tree_leaf_number-diff,
			startDepth, next_left_leaf_number_subtree,
			space+1, m)
	} else {
		proofs = append(proofs, &ProofElem{
			Cat:     1,
			Right:   false,
			LeafNum: 0,
			Res: &proofRes{
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
		proofs = generateProofRecursive(right_node, right, proofs,
			max_left_tree_leaf_number+diff, startDepth,
			leaf_number_sub_tree-next_left_leaf_number_subtree,
			space+1, m)
	} else {
		proofs = append(proofs, &ProofElem{
			Cat:     1,
			Right:   true,
			LeafNum: 0,
			Res: &proofRes{
				h:  right_node.getHash(),
				td: right_node.getDifficulty(),
			},
		})
	}
	return proofs
}

func (m *mmr) genProof(right_difficulty *big.Int, blocks []uint64) *ProofInfo {
	blocks = SortAndRemoveRepeatForBlocks(blocks)
	proofs, rootNode, depth := []*ProofElem{}, m.getRootNode(), get_depth(m.getLeafNumber())
	max_leaf_num := uint64(math.Pow(float64(2), float64(depth-1)))
	proofs = generateProofRecursive(rootNode, blocks, proofs, max_leaf_num, depth,
		m.getLeafNumber(), 0, m)

	proofs = append(proofs, &ProofElem{
		Cat:     0,
		Right:   false,
		LeafNum: m.getLeafNumber(),
		Res: &proofRes{
			h:  rootNode.getHash(),
			td: rootNode.getDifficulty(),
		},
	})
	return &ProofInfo{
		RootHash:       m.getRoot(),
		RootDifficulty: m.getRootDifficulty(),
		LeafNumber:     m.getLeafNumber(),
		Elems:          proofs,
	}
}

func (m *mmr) CreateNewProof(right_difficulty *big.Int) (*ProofInfo, []uint64, []uint64) {
	root_hash := m.getRoot()
	r1, _ := new(big.Float).SetInt(right_difficulty).Float64()
	r2, _ := new(big.Float).SetInt(new(big.Int).Add(m.getRootDifficulty(), right_difficulty)).Float64()
	required_queries := uint64(vd_calculate_m(float64(lambda), c, r1, r2, m.getLeafNumber()) + 1.0)

	weights, blocks := []float64{}, []uint64{}
	for i := 0; i < int(required_queries); i++ {
		h := RlpHash([]interface{}{root_hash, uint64(i)})
		random := Hash_to_f64(h)
		r3, _ := new(big.Float).SetInt(m.getRootDifficulty()).Float64()
		AggrWeight := cdf(random, vd_calculate_delta(r1, r3))
		weights = append(weights, AggrWeight)
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
	return m.genProof(right_difficulty, blocks), blocks, extra_blocks
}

///////////////////////////////////////////////////////////////////////////////////////

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
			hash := merge2(node1.Res.h, node2.Res.h)
			tmp_nodes = append(tmp_nodes, &VerifyElem{
				Res: &proofRes{
					h:  hash,
					td: new(big.Int).Add(node1.Res.td, node2.Res.td),
				},
				Index:      math.MaxUint64, // uint64(-1) .. none
				LeafNumber: math.MaxUint64, // uint64(-1) .. none
			})
		} else {
			break
		}
	}
	if len(tmp_nodes) >= 1 {
		return tmp_nodes[0].Res.h, tmp_nodes[0].Res.td
	}
	return common.Hash{0}, nil
}
func (p *ProofInfo) VerifyProof(blocks []*ProofBlock) bool {
	blocks = SortAndRemoveRepeatForProofBlocks(blocks)
	blocks = reverseForProofBlocks(blocks)
	proof_blocks := ProofBlocks(blocks)

	proofs := ProofElems(p.Elems)
	root_elem := proofs.pop_back()
	if root_elem == nil || root_elem.Cat != 0 {
		return false
	}
	if len(proofs) == 1 {
		if it := proofs.pop_back(); it != nil {
			if it.Cat == 2 {
				return equal_hash(it.Res.h, root_elem.Res.h)
			}
		}
		return false
	}
	nodes := VerifyElems([]*VerifyElem{})
	for {
		if !proofs.is_empty() {
			proof_elem := proofs.pop_front()
			if proof_elem.Cat == 2 {
				proof_block := proof_blocks.pop()
				number := proof_block.Number

				if !nodes.is_empty() {
					//TODO: Verification of previous MMR should happen here
					//weil in einem Ethereum block header kein mmr hash vorhanden ist, kann man
					//dies nicht überprüfen, wenn doch irgendwann vorhanden, dann einfach
					//'block_header.mmr == old_root_hash' überprüfen
					_, left_difficulty := get_root(nodes)
					left, middle := new(big.Float).SetInt(left_difficulty), new(big.Float).Mul(new(big.Float).SetInt(root_elem.Res.td), big.NewFloat(proof_block.AggrWeight))
					right := new(big.Float).Add(new(big.Float).SetInt(left_difficulty), new(big.Float).SetInt(proof_elem.Res.td))
					if left.Cmp(middle) > 0 || right.Cmp(middle) <= 0 {
						// "aggregated difficulty is not correct, should coincide with: {} <= {} < {}",left, middle, right
						return false
					}
				}
				if number%2 == 0 && number != (root_elem.LeafNum-1) {
					right_node := proofs.pop_front()
					right_node_hash, right_node_diff := right_node.Res.h, new(big.Int).Set(right_node.Res.td)
					if right_node.Cat == 2 || right_node.Cat == 1 {
						if right_node.Cat == 2 {
							proof_blocks.pop()
						}
					} else {
						// Expected ???
						return false
					}
					hash := merge2(proof_elem.Res.h, right_node_hash)
					nodes = append(nodes, &VerifyElem{
						Res: &proofRes{
							h:  hash,
							td: new(big.Int).Add(proof_elem.Res.td, right_node_diff),
						},
						Index:      number / 2,
						LeafNumber: root_elem.LeafNum / 2,
					})
				} else {
					res0 := nodes.pop_back()
					hash := merge2(res0.Res.h, proof_elem.Res.h)
					nodes = append(nodes, &VerifyElem{
						Res: &proofRes{
							h:  hash,
							td: new(big.Int).Add(proof_elem.Res.td, res0.Res.td),
						},
						Index:      number / 2,
						LeafNumber: root_elem.LeafNum / 2,
					})
				}
			} else if proof_elem.Cat == 1 {
				if proof_elem.Right {
					left_node := nodes.pop_back()
					hash := merge2(left_node.Res.h, proof_elem.Res.h)
					nodes = append(nodes, &VerifyElem{
						Res: &proofRes{
							h:  hash,
							td: new(big.Int).Add(left_node.Res.td, proof_elem.Res.td),
						},
						Index:      left_node.Index / 2,
						LeafNumber: left_node.LeafNumber / 2,
					})
				} else {
					nodes = append(nodes, &VerifyElem{
						Res:        proof_elem.Res,
						Index:      math.MaxUint64, // UINT64(-1)
						LeafNumber: math.MaxUint64, // UINT64(-1)
					})
				}
			} else if proof_elem.Cat == 0 {
				// do nothing
			} else {
				panic("invalid Cat...")
			}
			for {
				if len(nodes) > 1 {
					node2 := nodes.pop_back()
					node1 := nodes.pop_back()
					if node2.Index%2 != 1 && !proofs.is_empty() {
						nodes = append(nodes, node1)
						nodes = append(nodes, node2)
						break
					}
					hash := merge2(node1.Res.h, node2.Res.h)
					nodes = append(nodes, &VerifyElem{
						Res: &proofRes{
							h:  hash,
							td: new(big.Int).Add(node1.Res.td, node2.Res.td),
						},
						Index:      node2.Index / 2,
						LeafNumber: node2.LeafNumber / 2,
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
		return equal_hash(root_elem.Res.h, res0.Res.h) && root_elem.Res.td.Cmp(res0.Res.td) == 0
	}
	return false
}
func VerifyRequiredBlocks(blocks []uint64, root_hash common.Hash, root_difficulty, right_difficulty *big.Int, root_leaf_number uint64) ([]*ProofBlock, error) {

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
		h := RlpHash([]interface{}{root_hash, uint64(i)})
		random := Hash_to_f64(h)
		r3, _ := new(big.Float).SetInt(root_difficulty).Float64()
		AggrWeight := cdf(random, vd_calculate_delta(r1, r3))
		weights = append(weights, AggrWeight)
	}
	sort.Float64s(weights)
	proof_blocks, weight_pos := []*ProofBlock{}, 0

	for _, v := range blocks {
		AggrWeight := weights[weight_pos]
		if len(extra_blocks) > 0 {
			index := len(extra_blocks) - 1
			curr_extra_block := extra_blocks[index]
			if v == curr_extra_block {
				extra_blocks = append(extra_blocks[:index], extra_blocks[index+1:]...)
				AggrWeight = 0 // 0--none
			} else {
				weight_pos++
			}
		} else {
			weight_pos++
		}
		proof_blocks = append(proof_blocks, &ProofBlock{
			Number:     v,
			AggrWeight: AggrWeight,
		})
	}
	return proof_blocks, nil
}
