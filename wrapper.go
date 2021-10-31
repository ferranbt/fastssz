package ssz

import "fmt"

// ProofTree hashes a HashRoot object with a Hasher from
// the default HasherPool
func ProofTree(v HashRoot) (*Node, error) {
	w := &Wrapper{}
	if err := v.HashTreeRootWith(w); err != nil {
		return nil, err
	}
	return w.Node(), nil
}

type Wrapper struct {
	nodes []*Node
}

/// --- wrapper implements the HashWalker interface ---

func (w *Wrapper) Index() int {
	return len(w.nodes)
}

func (w *Wrapper) Append(i []byte) {
	panic("TODO")
}

func (w *Wrapper) AppendUint64(i uint64) {
	panic("TODO")
}

func (w *Wrapper) AppendUint8(i uint8) {
	panic("TODO")
}

func (w *Wrapper) FillUpTo32() {
	panic("TODO")
}

func (w *Wrapper) Merkleize(indx int) {
	w.Commit(indx)
}

func (w *Wrapper) MerkleizeWithMixin(indx int, num, limit uint64) {
	w.CommitWithMixin(indx, int(num), int(limit))
}

func (w *Wrapper) PutBitlist(bb []byte, maxSize uint64) {

}

func (w *Wrapper) PutBool(b bool) {
	w.AddNode(LeafFromBool(b))
}

func (w *Wrapper) PutBytes(b []byte) {
	w.AddBytes(b)
}

func (w *Wrapper) PutUint64(i uint64) {

}

func (w *Wrapper) PutUint8(i uint8) {

}

/// --- legacy ones ---

func (w *Wrapper) AddBytes(b []byte) {

}

func (w *Wrapper) AddUint64(i uint64) {
	w.AddNode(LeafFromUint64(i))
}

func (w *Wrapper) AddUint32(i uint32) {
	w.AddNode(LeafFromUint32(i))
}

func (w *Wrapper) AddUint16(i uint16) {
	w.AddNode(LeafFromUint16(i))
}

func (w *Wrapper) AddUint8(i uint8) {
	w.AddNode(LeafFromUint8(i))
}

func (w *Wrapper) AddNode(n *Node) {
	if w.nodes == nil {
		w.nodes = []*Node{}
	}
	w.nodes = append(w.nodes, n)
}

func (w *Wrapper) Node() *Node {
	if len(w.nodes) != 1 {
		fmt.Println(w.nodes)
		panic("BAD")
	}
	return w.nodes[0]
}

func (w *Wrapper) Commit(i int) {
	res, err := TreeFromNodes(w.nodes[i:])
	if err != nil {
		panic(err)
	}
	// remove the old nodes
	w.nodes = w.nodes[:i]
	// add the new node
	w.AddNode(res)
}

func (w *Wrapper) CommitWithMixin(i, num, limit int) {
	res, err := TreeFromNodesWithMixin(w.nodes[i:], num, limit)
	if err != nil {
		panic(err)
	}
	// remove the old nodes
	w.nodes = w.nodes[:i]
	// add the new node
	w.AddNode(res)
}

func (w *Wrapper) AddEmpty() {
	w.AddNode(EmptyLeaf())
}
