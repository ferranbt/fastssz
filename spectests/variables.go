package spectests

var (
	// phase0
	historicalRoots   uint64
	eth1DataVotes     uint64
	epochAttestations uint64
	slashings         uint64
	randaoMixes       uint64
	rootsSize         uint64

	// altair
	syncCommitteeBits uint64
)

func init() {
	setMainnetSpec()
}

func setMainnetSpec() {
	historicalRoots = 8192
	eth1DataVotes = 2048
	epochAttestations = 4096
	slashings = 8192
	randaoMixes = 65536
	rootsSize = 8192
	syncCommitteeBits = 64
}

func setMinimalSpec() {
	historicalRoots = 64
	eth1DataVotes = 32
	epochAttestations = 1024
	slashings = 64
	randaoMixes = 64
	rootsSize = 64
	syncCommitteeBits = 4
}
