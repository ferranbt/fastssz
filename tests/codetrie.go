package tests

import (
	"encoding/binary"
	"errors"

	ssz "github.com/ferranbt/fastssz"
)

type Hash []byte

type Metadata struct {
	Version    uint8
	CodeHash   Hash `ssz-size:"32"`
	CodeLength uint16
}

type Chunk struct {
	FIO  uint8
	Code []byte `ssz-size:"32"` // Last chunk is right-padded with zeros
}

type CodeTrieSmall struct {
	Metadata *Metadata
	Chunks   []*Chunk `ssz-max:"4"`
}

type CodeTrieBig struct {
	Metadata *Metadata
	Chunks   []*Chunk `ssz-max:"1024"`
}

func (md *Metadata) GetTree() (*ssz.Node, error) {
	leaves := md.getLeaves()
	return ssz.TreeFromChunks(leaves)
}

func (md *Metadata) getLeaves() [][]byte {
	leaves := make([][]byte, 4)
	leaves[0] = ssz.LeafFromUint8(md.Version)
	leaves[1] = ssz.LeafFromBytes(md.CodeHash)
	leaves[2] = ssz.LeafFromUint16(md.CodeLength)
	leaves[3] = ssz.EmptyLeaf()
	return leaves
}

func (t *CodeTrieSmall) GetTree() (*ssz.Node, error) {
	// Metadata tree
	mdTree, err := t.Metadata.GetTree()
	if err != nil {
		return nil, err
	}
	chunkMixinTree, err := t.getChunkListTree()
	if err != nil {
		return nil, err
	}
	// Tree with metadata and chunks subtrees
	return ssz.NewNodeWithLR(mdTree, chunkMixinTree), nil
}

func (t *CodeTrieSmall) getChunkListTree() (*ssz.Node, error) {
	return getChunkListTree(4, t.Chunks)
}

func getChunkListTree(size int, chunks []*Chunk) (*ssz.Node, error) {
	// Construct a tree  for each chunk
	if len(chunks) > size {
		return nil, errors.New("Number of chunks exceeds capacity")
	}

	chunkTrees := make([]*ssz.Node, size)
	emptyLeaf := ssz.NewNodeWithValue(make([]byte, 32))
	for i := 0; i < size; i++ {
		chunkTrees[i] = emptyLeaf
		if i < len(chunks) {
			c := chunks[i]
			t, err := c.GetTree()
			if err != nil {
				return nil, err
			}
			chunkTrees[i] = t
		}
	}

	// Construct a tree out of all chunk subtrees
	chunksTree, err := ssz.TreeFromNodes(chunkTrees)
	if err != nil {
		return nil, err
	}

	// Mixin chunks len
	chunkCountLeafValue := make([]byte, 32)
	binary.LittleEndian.PutUint64(chunkCountLeafValue[:], uint64(len(chunks)))
	chunkCountLeaf := ssz.NewNodeWithValue(chunkCountLeafValue)

	chunkMixinTree := ssz.NewNodeWithLR(chunksTree, chunkCountLeaf)
	return chunkMixinTree, nil
}

func (t *CodeTrieBig) GetTree() (*ssz.Node, error) {
	// Metadata tree
	mdTree, err := t.Metadata.GetTree()
	if err != nil {
		return nil, err
	}
	chunkMixinTree, err := t.getChunkListTree()
	if err != nil {
		return nil, err
	}
	// Tree with metadata and chunks subtrees
	return ssz.NewNodeWithLR(mdTree, chunkMixinTree), nil
}

func (t *CodeTrieBig) getChunkListTree() (*ssz.Node, error) {
	return getChunkListTree(1024, t.Chunks)
}

func (c *Chunk) GetTree() (*ssz.Node, error) {
	leaves := c.getLeaves()
	return ssz.TreeFromChunks(leaves)
}

func (c *Chunk) getLeaves() [][]byte {
	leaves := make([][]byte, 2)
	leaves[0] = ssz.LeafFromUint8(c.FIO)
	leaves[1] = ssz.LeafFromBytes(c.Code)
	return leaves
}
