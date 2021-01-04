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
	chunks := make([][]byte, 4)
	chunks[0] = make([]byte, 32) // Version
	chunks[0][0] = md.Version
	chunks[1] = md.CodeHash[:]
	chunks[2] = make([]byte, 32)
	binary.LittleEndian.PutUint16(chunks[2][:2], md.CodeLength)
	chunks[3] = make([]byte, 32)
	return chunks
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
	chunks := make([][]byte, 2)
	chunks[0] = make([]byte, 32)
	chunks[0][0] = c.FIO
	chunks[1] = c.Code[:]
	return chunks
}
