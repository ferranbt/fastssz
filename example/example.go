package example

import "github.com/ferranbt/fastssz/example2"

type AttestationData struct {
	Slot            uint64               `json:"slot"`
	Index           uint64               `json:"index"`
	BeaconBlockHash []byte               `json:"beacon_block_root" ssz-size:"32"`
	Source          *example2.Checkpoint `json:"source"`
	Target          *example2.Checkpoint `json:"target"`
}
