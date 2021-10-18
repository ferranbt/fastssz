package ssz

import (
	"fmt"
	"math/bits"
)

type Wrapper struct {
	nodes []*Node
}

func (w *Wrapper) Indx() int {
	return len(w.nodes)
}

func (w *Wrapper) AddBytes(b []byte) {
	w.AddNode(LeafFromBytes(b))
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

func (w *Wrapper) AddBitlist(blist []byte, maxSize int) {
	tmp, size := parseBitlistForTree(blist)
	subIdx := w.Indx()
	w.AddBytes(tmp)
	w.CommitWithMixin(subIdx, int(size), (maxSize+255)/256)
}

func parseBitlistForTree(buf []byte) ([]byte, uint64) {
	dst := make([]byte, 0)
	msb := uint8(bits.Len8(buf[len(buf)-1])) - 1
	size := uint64(8*(len(buf)-1) + int(msb))

	dst = append(dst, buf...)
	dst[len(dst)-1] &^= uint8(1 << msb)

	newLen := len(dst)
	for i := len(dst) - 1; i >= 0; i-- {
		if dst[i] != 0x00 {
			break
		}
		newLen = i
	}
	res := dst[:newLen]
	return res, size
}
