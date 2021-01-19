package ssz

import "fmt"

type Wrapper struct {
	nodes []*Node
}

func (w *Wrapper) Indx() int {
	return len(w.nodes)
}

func (w *Wrapper) AddBytes(b []byte) {
	w.addNode(LeafFromBytes(b))
}

func (w *Wrapper) AddUint64(i uint64) {
	w.addNode(LeafFromUint64(i))
}

func (w *Wrapper) AddUint32(i uint32) {
	w.addNode(LeafFromUint32(i))
}

func (w *Wrapper) AddUint16(i uint16) {
	w.addNode(LeafFromUint16(i))
}

func (w *Wrapper) AddUint8(i uint8) {
	w.addNode(LeafFromUint8(i))
}

func (w *Wrapper) addNode(n *Node) {
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
	w.addNode(res)
}

func (w *Wrapper) AddEmpty() {
	w.addNode(EmptyLeaf())
}
