package altair

type SyncCommitteeDuty struct {
	Pubkey               []byte `json:"pubkey,omitempty" ssz-size:"48"`
	ValidatorIndex       uint64 `json:"validator_index,omitempty"`
	SyncCommitteeIndices uint64 `json:"sync_committee_indices,omitempty"`
}
