package spectests

var (
	eth1DataVotes     uint64
	epochAttestations uint64
	slashings         uint64
	randaoMixes       uint64
	rootsSize         uint64
)

func init() {
	setMainnetSpec()
}

func setMinimalSpec() {
	eth1DataVotes = 32
	epochAttestations = 1024
	slashings = 64
	randaoMixes = 64
	rootsSize = 64
}

func setMainnetSpec() {
	eth1DataVotes = 2048
	epochAttestations = 4096
	slashings = 8192
	randaoMixes = 65536
	rootsSize = 8192
}
