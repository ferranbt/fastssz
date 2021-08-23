package minimal

import (
	external2Alias "github.com/ferranbt/fastssz/spectests/external2"
)

type BeaconState struct {
	GenesisTime       uint64             `json:"genesis_time"`
	GenesisValidatorsRoot       []byte   `json:"genesis_validators_root,omitempty" ssz-size:"32"`
	Slot              uint64             `json:"slot"`
	Fork              *Fork              `json:"fork"`
	LatestBlockHeader *BeaconBlockHeader `json:"latest_block_header"`
	// BlockRoots is 8192,32 in mainnet
	BlockRoots        [][]byte       `json:"block_roots" ssz-size:"64,32"`
	// StateRoots is 8192,32 in mainnet
	StateRoots        [][]byte         `json:"state_roots" ssz-size:"64,32"`
	HistoricalRoots   [][]byte         `json:"historical_roots" ssz-max:"16777216" ssz-size:"?,32"`
	Eth1Data          *Eth1Data          `json:"eth1_data"`
	// Eth1DataVotes is 2048 in mainnet
	Eth1DataVotes     []*Eth1Data        `json:"eth1_data_votes" ssz-max:"32"`
	Eth1DepositIndex  uint64             `json:"eth1_deposit_index"`
	Validators        []*Validator       `json:"validators" ssz-max:"1099511627776"`
	Balances          []uint64           `json:"balances" ssz-max:"1099511627776"`
	RandaoMixes       [][]byte           `json:"randao_mixes" ssz-size:"64,32"`
	Slashings         []uint64           `json:"slashings" ssz-size:"64"`

	// PreviousEpochAttestations is 4096 in mainnet
	PreviousEpochAttestations []*PendingAttestation `json:"previous_epoch_attestations" ssz-max:"1024"`
	// CurrentEpochAttestations is 4096 in mainnet
	CurrentEpochAttestations  []*PendingAttestation `json:"current_epoch_attestations" ssz-max:"1024"`
	JustificationBits         []byte                `json:"justification_bits" ssz-size:"1"`

	PreviousJustifiedCheckpoint *Checkpoint `json:"previous_justified_checkpoint"`
	CurrentJustifiedCheckpoint  *Checkpoint `json:"current_justified_checkpoint"`
	FinalizedCheckpoint         *Checkpoint `json:"finalized_checkpoint"`
}

type HistoricalBatch struct {
	// BlockRoots is 8192,32 in mainnet
	BlockRoots [][]byte `json:"block_roots,omitempty" ssz-size:"64,32"`
	// StateRoots is 8192,32 in mainnet
	StateRoots [][]byte `json:"state_roots,omitempty" ssz-size:"64,32"`
}

// ALL THE TYPES BELOW THIS LINE ARE UNCHANGED, THEY HAVE BEEN COPIED TO WORK AROUND
// FASTSSZ CODEGEN BUGS

type Eth1Data struct {
	DepositRoot  []byte `json:"deposit_root" ssz-size:"32"`
	DepositCount uint64 `json:"deposit_count"`
	BlockHash    []byte `json:"block_hash" ssz-size:"32"`
}

type Validator struct {
	Pubkey                     []byte `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials      []byte `json:"withdrawal_credentials" ssz-size:"32"`
	EffectiveBalance           uint64 `json:"effective_balance"`
	Slashed                    bool   `json:"slashed"`
	ActivationEligibilityEpoch uint64 `json:"activation_eligibility_epoch"`
	ActivationEpoch            uint64 `json:"activation_epoch"`
	ExitEpoch                  uint64 `json:"exit_epoch"`
	WithdrawableEpoch          uint64 `json:"withdrawable_epoch"`
}
type PendingAttestation struct {
	AggregationBits []byte           `json:"aggregation_bits" ssz:"bitlist" ssz-max:"2048"`
	Data            *AttestationData `json:"data"`
	InclusionDelay  uint64           `json:"inclusion_delay"`
	ProposerIndex   uint64           `json:"proposer_index"`
}

type Fork struct {
	PreviousVersion []byte `json:"previous_version" ssz-size:"4"`
	CurrentVersion  []byte `json:"current_version" ssz-size:"4"`
	Epoch           uint64 `json:"epoch"`
}

type BeaconBlockHeader struct {
	Slot       uint64 `json:"slot"`
	ProposerIndex uint64 `json:"proposer_index"`
	ParentRoot []byte `json:"parent_root" ssz-size:"32"`
	StateRoot  []byte `json:"state_root" ssz-size:"32"`
	BodyRoot   []byte `json:"body_root" ssz-size:"32"`
}

type Checkpoint struct {
	Epoch external2Alias.EpochAlias `json:"epoch"`
	Root  []byte                    `json:"root" ssz-size:"32"`
}

type Slot uint64 // alias from the same package

type AttestationData struct {
	Slot            Slot        `json:"slot"`
	Index           uint64      `json:"index"`
	BeaconBlockRoot []byte `json:"beacon_block_root" ssz-size:"32"`
	Source          *Checkpoint `json:"source"`
	Target          *Checkpoint `json:"target"`
}