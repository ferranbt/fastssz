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

func TreeFromChunks(chunks [][]byte) (*Node, error) {
	numLeaves := len(chunks)
	if !isPowerOfTwo(numLeaves) {
		return nil, errors.New("Number of leaves should be a power of 2")
	}

	numNodes := numLeaves*2 - 1
	nodes := make([]*Node, numNodes)
	for i := numNodes; i > 0; i-- {
		// Is a leaf
		if i > numNodes-numLeaves {
			val := chunks[i-numLeaves]
			nodes[i-1] = &Node{left: nil, right: nil, value: val}
		} else {
			// Is a branch node
			nodes[i-1] = &Node{left: nodes[(i*2)-1], right: nodes[(i*2+1)-1], value: nil}
		}
	}

	return nodes[0], nil
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

func (n *Node) Hash() []byte {
	// TODO: handle special cases: empty root, one non-empty node
	return hashNode(n)
}

func hashNode(n *Node) []byte {
	// Leaf
	if n.left == nil && n.right == nil {
		return n.value
	}
	// Only one child
	if n.left == nil || n.right == nil {
		panic("Tree incomplete")
	}
	return hashFn(append(hashNode(n.left), hashNode(n.right)...))
}

func isPowerOfTwo(n int) bool {
	return (n & (n - 1)) == 0
}
