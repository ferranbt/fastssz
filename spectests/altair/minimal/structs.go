package altair

import (
	"github.com/ferranbt/fastssz/spectests/external"
	external2Alias "github.com/ferranbt/fastssz/spectests/external2"
)

// NEW TYPES

type ContributionAndProof struct {
	AggregatorIndex uint64 `json:"aggregator_index,omitempty"`
	Contribution    *SyncCommitteeContribution                         `protobuf:"bytes,2,opt,name=contribution,proto3" json:"contribution,omitempty"`
	SelectionProof  []byte                                             `protobuf:"bytes,3,opt,name=selection_proof,json=selectionProof,proto3" json:"selection_proof,omitempty" ssz-size:"96"`
}

type SyncAggregate struct {
	SyncCommitteeBits      []byte `json:"sync_committee_bits,omitempty" ssz-size:"4"`
	SyncCommitteeSignature []byte `json:"sync_committee_signature,omitempty" ssz-size:"96"`
}

type SyncAggregatorSelectionData struct {
	Slot              uint64 `json:"slot,omitempty"`
	SubcommitteeIndex uint64 `json:"subcommittee_index,omitempty"`
}

type SyncCommittee struct {
	Pubkeys         [][]byte `json:"pubkeys,omitempty" ssz-size:"32,48"`
	AggregatePubkey []byte   `json:"aggregate_pubkey,omitempty" ssz-size:"48"`
}

type SyncCommitteeDuty struct {
	Pubkey               []byte `json:"pubkey,omitempty" ssz-size:"48"`
	ValidatorIndex       uint64 `json:"validator_index,omitempty"`
	SyncCommitteeIndices uint64 `json:"sync_committee_indices,omitempty"`
}

type SyncCommitteeContribution struct {
	Slot              uint64 `json:"slot,omitempty"`
	BlockRoot         []byte `json:"block_root,omitempty" ssz-size:"32"`
	SubcommitteeIndex uint64 `json:"subcommittee_index,omitempty"`
	AggregationBits   []byte `json:"aggregation_bits,omitempty" ssz-size:"1"`
	Signature         []byte `json:"signature,omitempty" ssz-size:"96"`
}

type SyncCommitteeMessage struct {
	Slot              uint64 `json:"slot,omitempty"`
	BlockRoot         []byte `json:"block_root,omitempty" ssz-size:"32"`
	ValidatorIndex uint64 `json:"validator_index,omitempty"`
	Signature      []byte `json:"signature,omitempty" ssz-size:"96"`
}

type SignedContributionAndProof struct {
	Message   *ContributionAndProof `json:"message,omitempty"`
	Signature []byte                `json:"signature,omitempty" ssz-size:"96"`
}

type LightClientSnapshot struct {
	Header *BeaconBlockHeader `json:"header"`
	CurrentSyncCommittee *SyncCommittee `json:"current_sync_committee"`
	NextSyncCommittee *SyncCommittee `json:"next_sync_committee"`
}

// TODO: this type will take a little more work to translate, putting that off for now
type LightClientUpdate struct {
	Header *BeaconBlockHeader `json:"header"`
	NextSyncCommittee *SyncCommittee `json:"next_sync_committee"`
	// TODO: figure out how the size of the next_sync_committee_branch Vector should be interpreted
	// next_sync_committee_branch: Vector[Bytes32, floorlog2(NEXT_SYNC_COMMITTEE_INDEX)]
	FinalityHeader *BeaconBlockHeader `json:"finality_header"`
	// TODO: figure out how the size of the finality_branch Vector should be interpreted
	// finality_branch: Vector[Bytes32, floorlog2(FINALIZED_ROOT_INDEX)]
}

// CHANGED TYPES

type BeaconBlockBody struct {
	RandaoReveal      []byte                 `json:"randao_reveal" ssz-size:"96"`
	Eth1Data          *Eth1Data              `json:"eth1_data"`
	Graffiti          []byte               `json:"graffiti" ssz-size:"32"`
	ProposerSlashings []*ProposerSlashing    `json:"proposer_slashings" ssz-max:"16"`
	AttesterSlashings []*AttesterSlashing    `json:"attester_slashings" ssz-max:"2"`
	Attestations      []*Attestation         `json:"attestations" ssz-max:"128"`
	Deposits          []*Deposit             `json:"deposits" ssz-max:"16"`
	VoluntaryExits    []*SignedVoluntaryExit `json:"voluntary_exits" ssz-max:"16"`
	SyncAggregate *SyncAggregate `json:"sync_aggregate"`
}

type BeaconState struct {
	GenesisTime       uint64             `json:"genesis_time"`
	GenesisValidatorsRoot       []byte   `json:"genesis_validators_root,omitempty" ssz-size:"32"`
	Slot              uint64             `json:"slot"`
	Fork              *Fork              `json:"fork"`
	LatestBlockHeader *BeaconBlockHeader `json:"latest_block_header"`
	BlockRoots        [][]byte       `json:"block_roots" ssz-size:"64,32"`
	StateRoots        [][]byte         `json:"state_roots" ssz-size:"64,32"`
	HistoricalRoots   [][]byte         `json:"historical_roots" ssz-max:"16777216" ssz-size:"?,32"`
	Eth1Data          *Eth1Data          `json:"eth1_data"`
	Eth1DataVotes     []*Eth1Data        `json:"eth1_data_votes" ssz-max:"32"`
	Eth1DepositIndex  uint64             `json:"eth1_deposit_index"`
	Validators        []*Validator       `json:"validators" ssz-max:"1099511627776"`
	Balances          []uint64           `json:"balances" ssz-max:"1099511627776"`
	RandaoMixes       [][]byte           `json:"randao_mixes" ssz-size:"64,32"`
	Slashings         []uint64           `json:"slashings" ssz-size:"64"`

	// modified in altair
	PreviousEpochParticipation  []byte `json:"previous_epoch_participation,omitempty" ssz-max:"1099511627776"`
	// modified in altair
	CurrentEpochParticipation   []byte `json:"current_epoch_participation,omitempty" ssz-max:"1099511627776"`
	JustificationBits         []byte   `json:"justification_bits" ssz-size:"1"`

	PreviousJustifiedCheckpoint *Checkpoint `json:"previous_justified_checkpoint"`
	CurrentJustifiedCheckpoint  *Checkpoint `json:"current_justified_checkpoint"`
	FinalizedCheckpoint         *Checkpoint `json:"finalized_checkpoint"`

	// new in altair
	InactivityScores            []uint64 `json:"inactivity_scores,omitempty" ssz-max:"1099511627776"`
	// new in altair
	CurrentSyncCommittee        *SyncCommittee `json:"current_sync_committee,omitempty"`
	// new in altair
	NextSyncCommittee           *SyncCommittee `json:"next_sync_committee,omitempty"`
}

// changed because BeaconBlockBody changed
type BeaconBlock struct {
	Slot       uint64           `json:"slot"`
	ProposerIndex uint64 `json:"proposer_index,omitempty"`
	ParentRoot []byte           `json:"parent_root" ssz-size:"32"`
	StateRoot  []byte           `json:"state_root" ssz-size:"32"`
	Body       *BeaconBlockBody `json:"body"`
}

type SignedBeaconBlock struct {
	Block     *BeaconBlock `json:"message"`
	Signature []byte       `json:"signature" ssz-size:"96"`
}

// TYPES THAT WE HAVE TO COPY/PASTE BECAUSE OF CODE GENERATION BUGS
// (references to the other spectest packages either panic (lists of references)
// or generate incorrect code (Eth1Data was treated as a uint for HTR)

type Attestation struct {
	AggregationBits []byte              `json:"aggregation_bits" ssz:"bitlist" ssz-max:"2048"`
	Data            *AttestationData    `json:"data"`
	Signature       *external.Signature `json:"signature" ssz-size:"96"`
}

type AttestationData struct {
	Slot            Slot        `json:"slot"`
	Index           uint64      `json:"index"`
	BeaconBlockRoot []byte `json:"beacon_block_root" ssz-size:"32"`
	Source          *Checkpoint `json:"source"`
	Target          *Checkpoint `json:"target"`
}

type AttesterSlashing struct {
	Attestation1 *IndexedAttestation `json:"attestation_1"`
	Attestation2 *IndexedAttestation `json:"attestation_2"`
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

type Deposit struct {
	Proof [][]byte `ssz-size:"33,32"`
	Data  *DepositData
}

type DepositData struct {
	Pubkey                [48]byte       `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials [32]byte       `json:"withdrawal_credentials" ssz-size:"32"`
	Amount                uint64         `json:"amount"`
	Signature             external.Bytes `json:"signature" ssz-size:"96"`
	Root                  [32]byte       `ssz:"-"`
}

type Eth1Data struct {
	DepositRoot  []byte `json:"deposit_root" ssz-size:"32"`
	DepositCount uint64 `json:"deposit_count"`
	BlockHash    []byte `json:"block_hash" ssz-size:"32"`
}

type Fork struct {
	PreviousVersion []byte `json:"previous_version" ssz-size:"4"`
	CurrentVersion  []byte `json:"current_version" ssz-size:"4"`
	Epoch           uint64 `json:"epoch"`
}

type IndexedAttestation struct {
	AttestationIndices []uint64         `json:"attesting_indices" ssz-max:"2048"`
	Data               *AttestationData `json:"data"`
	Signature          []byte           `json:"signature" ssz-size:"96"`
}

type PendingAttestation struct {
	AggregationBits []byte           `json:"aggregation_bits" ssz:"bitlist" ssz-max:"2048"`
	Data            *AttestationData `json:"data"`
	InclusionDelay  uint64           `json:"inclusion_delay"`
	ProposerIndex   uint64           `json:"proposer_index"`
}

type ProposerSlashing struct {
	Header1       *SignedBeaconBlockHeader `json:"signed_header_1"`
	Header2       *SignedBeaconBlockHeader `json:"signed_header_2"`
}

type SignedBeaconBlockHeader struct {
	Header    *BeaconBlockHeader `json:"message"`
	Signature []byte             `json:"signature" ssz-size:"96"`
}

type SignedVoluntaryExit struct {
	Exit      *VoluntaryExit          `json:"message"`
	Signature external.FixedSignature `json:"signature" ssz-size:"96"`
}

type Slot uint64 // alias from the same package

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

type VoluntaryExit struct {
	Epoch          uint64 `json:"epoch"`
	ValidatorIndex uint64 `json:"validator_index"`
}
