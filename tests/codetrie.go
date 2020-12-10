package tests

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
