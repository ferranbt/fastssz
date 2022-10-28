package other

import ssz "github.com/ferranbt/fastssz"

type Case4Interface struct {
}

func (c *Case4Interface) SizeSSZ() (size int) {
	return 96
}

func (s *Case4Interface) MarshalSSZTo(buf []byte) ([]byte, error) {
	return nil, nil
}

func (s *Case4Interface) HashTreeRootWith(hh ssz.HashWalker) (err error) {
	return
}

func (s *Case4Interface) UnmarshalSSZ(buf []byte) error {
	return nil
}

type Case4FixedSignature [96]byte

type Case4Bytes []byte
