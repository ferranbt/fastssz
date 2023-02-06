package spectests

import (
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	ssz "github.com/ferranbt/fastssz"
	"github.com/stretchr/testify/require"
)

const TestFileName = "fixtures/beacon_state_bellatrix.ssz" // https://goerli.beaconcha.in/slot/4744352

func TestBeaconHeader_SingleProof(t *testing.T) {
	data, err := os.ReadFile(TestFileName)
	require.NoError(t, err)

	sszState := BeaconStateBellatrix{}
	err = sszState.UnmarshalSSZ(data)
	require.NoError(t, err)

	object := sszState.LatestBlockHeader

	objectTree, err := object.GetTree()
	require.NoError(t, err)

	proofAtIndex := 8

	proof, err := objectTree.Prove(proofAtIndex)
	require.NoError(t, err)

	require.Equal(t, hex.EncodeToString(proof.Leaf), "a064480000000000000000000000000000000000000000000000000000000000")
	require.Equal(t, proofAtIndex, proof.Index)
	require.Equal(t, 3, len(proof.Hashes), "proof hashes length incorrect")
	require.Equal(t, hex.EncodeToString(proof.Hashes[0]), "7859010000000000000000000000000000000000000000000000000000000000")
	require.Equal(t, hex.EncodeToString(proof.Hashes[1]), "007c0d1e0260fb9a6fa86a39569aaebc9a95aaab0180f2865da2fc25180e2242")
	require.Equal(t, hex.EncodeToString(proof.Hashes[2]), "98a517b0aa099cdfd06cdcfe71869417a00f168e8bec03ee3fa2135af2396bb6")
}

func TestBeaconHeader_MultiProof(t *testing.T) {
	data, err := os.ReadFile(TestFileName)
	require.NoError(t, err)

	sszState := BeaconStateBellatrix{}
	err = sszState.UnmarshalSSZ(data)
	require.NoError(t, err)

	object := sszState.LatestBlockHeader

	objectTree, err := object.GetTree()
	require.NoError(t, err)

	proofAtIndices := []int{8, 9, 13}

	multiProof, err := objectTree.ProveMulti(proofAtIndices)
	require.NoError(t, err)

	require.Equal(t, 3, len(multiProof.Leaves), "multi proof leaf hashes length incorrect")
	require.Equal(t, hex.EncodeToString(multiProof.Leaves[0]), "a064480000000000000000000000000000000000000000000000000000000000")
	require.Equal(t, hex.EncodeToString(multiProof.Leaves[1]), "7859010000000000000000000000000000000000000000000000000000000000")
	require.Equal(t, hex.EncodeToString(multiProof.Leaves[2]), "0000000000000000000000000000000000000000000000000000000000000000")

	require.Equal(t, proofAtIndices, multiProof.Indices)
	require.Equal(t, 3, len(multiProof.Hashes), "proof hashes length incorrect")
	require.Equal(t, hex.EncodeToString(multiProof.Hashes[0]), "445fab586d7d52993d7713c29da316d7e0fe04fd053983198af93fb131ce02ed")
	require.Equal(t, hex.EncodeToString(multiProof.Hashes[1]), "f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b")
	require.Equal(t, hex.EncodeToString(multiProof.Hashes[2]), "007c0d1e0260fb9a6fa86a39569aaebc9a95aaab0180f2865da2fc25180e2242")
}

func TestBeaconState_BlockRootAtIndexProof(t *testing.T) {
	data, err := os.ReadFile(TestFileName)
	require.NoError(t, err)

	sszState := BeaconStateBellatrix{}
	err = sszState.UnmarshalSSZ(data)
	require.NoError(t, err)

	// index of first block_root field in the beacon state
	leavesStart := 303104
	// let's prove block roof at position 4 in the block roots array
	index := leavesStart + 3

	expectedLeaf := sszState.BlockRoots[3]

	tree, err := sszState.GetTree()
	require.NoError(t, err)

	proof, err := tree.Prove(index)
	require.NoError(t, err)

	// check that the block root hash at the index matches what is in the beacon state
	require.Equal(t, expectedLeaf, proof.Leaf)

	root, err := sszState.HashTreeRoot()

	ok, err := ssz.VerifyProof(root[:], proof)
	require.NoError(t, err, "failed to verify proof")
	require.True(t, ok, "incorrect proof verification")
}

func TestBeaconState_BlockRootsProof(t *testing.T) {
	data, err := os.ReadFile(TestFileName)
	require.NoError(t, err)

	sszState := BeaconStateBellatrix{}
	err = sszState.UnmarshalSSZ(data)
	require.NoError(t, err)

	index := 37

	tree, err := sszState.GetTree()
	require.NoError(t, err)

	root, err := sszState.HashTreeRoot()

	// This is required to set the node values as the tree is hashed. Ideally should be done as part of GetTree() or Prove()
	tree.Hash()

	proof, err := tree.Prove(index)
	require.NoError(t, err)

	ok, err := ssz.VerifyProof(root[:], proof)
	require.NoError(t, err, "failed to verify proof")
	require.True(t, ok, "incorrect proof verification")
}

func TestBeaconStateTree_HashTreeRoot(t *testing.T) {
	data, err := os.ReadFile(TestFileName)
	require.NoError(t, err)

	sszState := BeaconStateBellatrix{}
	err = sszState.UnmarshalSSZ(data)
	require.NoError(t, err)

	fmt.Println(sszState.Slot)

	tree, err := sszState.GetTree()
	require.NoError(t, err)

	hash := tree.Hash()

	// taken from https://goerli.beaconcha.in/slot/4744352 - stateRoot field
	require.Equal(t, "c4a9c5ebf637c089db599574b568bb679b385c1984f08410707db08e03d7ae52", hex.EncodeToString(hash))
}
