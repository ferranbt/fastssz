package spectests2

import (
	v1alpha1 "github.com/ferranbt/fastssz/spectests2/aux"
	github_com_prysmaticlabs_go_bitfield "github.com/prysmaticlabs/go-bitfield"
)

type BeaconState struct {
	GenesisTime                 uint64                                          `protobuf:"varint,1001,opt,name=genesis_time,json=genesisTime,proto3" json:"genesis_time,omitempty"`
	Slot                        uint64                                          `protobuf:"varint,1002,opt,name=slot,proto3" json:"slot,omitempty"`
	Fork                        *Fork                                           `protobuf:"bytes,1003,opt,name=fork,proto3" json:"fork,omitempty"`
	LatestBlockHeader           *v1alpha1.BeaconBlockHeader                     `protobuf:"bytes,2001,opt,name=latest_block_header,json=latestBlockHeader,proto3" json:"latest_block_header,omitempty"`
	BlockRoots                  [][]byte                                        `protobuf:"bytes,2002,rep,name=block_roots,json=blockRoots,proto3" json:"block_roots,omitempty" ssz-size:"8192,32"`
	StateRoots                  [][]byte                                        `protobuf:"bytes,2003,rep,name=state_roots,json=stateRoots,proto3" json:"state_roots,omitempty" ssz-size:"8192,32"`
	HistoricalRoots             [][]byte                                        `protobuf:"bytes,2004,rep,name=historical_roots,json=historicalRoots,proto3" json:"historical_roots,omitempty" ssz-size:"?,32" ssz-max:"16777216"`
	Eth1Data                    *v1alpha1.Eth1Data                              `protobuf:"bytes,3001,opt,name=eth1_data,json=eth1Data,proto3" json:"eth1_data,omitempty"`
	Eth1DataVotes               []*v1alpha1.Eth1Data                            `protobuf:"bytes,3002,rep,name=eth1_data_votes,json=eth1DataVotes,proto3" json:"eth1_data_votes,omitempty" ssz-max:"1024"`
	Eth1DepositIndex            uint64                                          `protobuf:"varint,3003,opt,name=eth1_deposit_index,json=eth1DepositIndex,proto3" json:"eth1_deposit_index,omitempty"`
	Validators                  []*v1alpha1.Validator                           `protobuf:"bytes,4001,rep,name=validators,proto3" json:"validators,omitempty" ssz-max:"1099511627776"`
	Balances                    []uint64                                        `protobuf:"varint,4002,rep,packed,name=balances,proto3" json:"balances,omitempty" ssz-max:"1099511627776"`
	RandaoMixes                 [][]byte                                        `protobuf:"bytes,5001,rep,name=randao_mixes,json=randaoMixes,proto3" json:"randao_mixes,omitempty" ssz-size:"65536,32"`
	Slashings                   []uint64                                        `protobuf:"varint,6001,rep,packed,name=slashings,proto3" json:"slashings,omitempty" ssz-size:"8192"`
	PreviousEpochAttestations   []*v1alpha1.PendingAttestation                  `protobuf:"bytes,7001,rep,name=previous_epoch_attestations,json=previousEpochAttestations,proto3" json:"previous_epoch_attestations,omitempty" ssz-max:"4096"`
	CurrentEpochAttestations    []*v1alpha1.PendingAttestation                  `protobuf:"bytes,7002,rep,name=current_epoch_attestations,json=currentEpochAttestations,proto3" json:"current_epoch_attestations,omitempty" ssz-max:"4096"`
	JustificationBits           github_com_prysmaticlabs_go_bitfield.Bitvector4 `protobuf:"bytes,8001,opt,name=justification_bits,json=justificationBits,proto3,casttype=github.com/prysmaticlabs/go-bitfield.Bitvector4" json:"justification_bits,omitempty" ssz-size:"1"`
	PreviousJustifiedCheckpoint *v1alpha1.Checkpoint                            `protobuf:"bytes,8002,opt,name=previous_justified_checkpoint,json=previousJustifiedCheckpoint,proto3" json:"previous_justified_checkpoint,omitempty"`
	CurrentJustifiedCheckpoint  *v1alpha1.Checkpoint                            `protobuf:"bytes,8003,opt,name=current_justified_checkpoint,json=currentJustifiedCheckpoint,proto3" json:"current_justified_checkpoint,omitempty"`
	FinalizedCheckpoint         *v1alpha1.Checkpoint                            `protobuf:"bytes,8004,opt,name=finalized_checkpoint,json=finalizedCheckpoint,proto3" json:"finalized_checkpoint,omitempty"`
}
