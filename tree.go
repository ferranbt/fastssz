package ssz

import (
	"errors"
)

type Node struct {
	left  *Node
	right *Node

	value []byte
}

func NewTree() *Node {
	return &Node{left: nil, right: nil, value: nil}
}

func TreeFromChunks(chunks [][]byte) *Node {
	numLeaves := int(nextPowerOfTwo(uint64(len(chunks))))
	numEmpty := numLeaves - len(chunks)
	numNodes := numLeaves*2 - 1

	nodes := make([]*Node, numNodes)
	for i := numNodes; i > 0; i-- {
		// It's a leaf
		if i > numNodes-numLeaves {
			val := []byte(nil)
			if i <= numNodes-numEmpty {
				val = chunks[i-numLeaves]
			}
			nodes[i-1] = &Node{left: nil, right: nil, value: val}
		} else {
			nodes[i-1] = &Node{left: nodes[(i*2)-1], right: nodes[(i*2+1)-1], value: nil}
		}
	}

	return nodes[0]
}

func (n *Node) Get(index int) (*Node, error) {
	pathLen := getPathLength(index)
	cur := n
	for i := pathLen - 1; i >= 0; i-- {
		if isRight := getPosAtLevel(index, i); isRight {
			cur = cur.right
		} else {
			cur = cur.left
		}
		if cur == nil {
			return nil, errors.New("Node not found in tree")
		}
	}

	return cur, nil
}
