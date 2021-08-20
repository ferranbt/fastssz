package altair

import ssz "github.com/ferranbt/fastssz"

// MarshalSSZ ssz marshals the SyncCommitteeDuty object
func (s *SyncCommitteeDuty) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(s)
}

// MarshalSSZTo ssz marshals the SyncCommitteeDuty object to a target array
func (s *SyncCommitteeDuty) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf

	// Field (0) 'Pubkey'
	if len(s.Pubkey) != 48 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, s.Pubkey...)

	// Field (1) 'ValidatorIndex'
	dst = ssz.MarshalUint64(dst, s.ValidatorIndex)

	// Field (2) 'SyncCommitteeIndices'
	dst = ssz.MarshalUint64(dst, s.SyncCommitteeIndices)

	return
}

// UnmarshalSSZ ssz unmarshals the SyncCommitteeDuty object
func (s *SyncCommitteeDuty) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 64 {
		return ssz.ErrSize
	}

	// Field (0) 'Pubkey'
	if cap(s.Pubkey) == 0 {
		s.Pubkey = make([]byte, 0, len(buf[0:48]))
	}
	s.Pubkey = append(s.Pubkey, buf[0:48]...)

	// Field (1) 'ValidatorIndex'
	s.ValidatorIndex = ssz.UnmarshallUint64(buf[48:56])

	// Field (2) 'SyncCommitteeIndices'
	s.SyncCommitteeIndices = ssz.UnmarshallUint64(buf[56:64])

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the SyncCommitteeDuty object
func (s *SyncCommitteeDuty) SizeSSZ() (size int) {
	size = 64
	return
}

// HashTreeRoot ssz hashes the SyncCommitteeDuty object
func (s *SyncCommitteeDuty) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(s)
}

// HashTreeRootWith ssz hashes the SyncCommitteeDuty object with a hasher
func (s *SyncCommitteeDuty) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Pubkey'
	if len(s.Pubkey) != 48 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(s.Pubkey)

	// Field (1) 'ValidatorIndex'
	hh.PutUint64(s.ValidatorIndex)

	// Field (2) 'SyncCommitteeIndices'
	hh.PutUint64(s.SyncCommitteeIndices)

	hh.Merkleize(indx)
	return
}
