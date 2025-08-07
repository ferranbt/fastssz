package spectests

var (
	ethDataVotes      uint32
	epochAttestations uint32
	slashings         uint32
	randaoMixes       uint32
	rootsSize         uint32
)

func setMinimalSpec() {
	ethDataVotes = 32
	epochAttestations = 1024
	slashings = 64
	randaoMixes = 64
	rootsSize = 64
}

func setMainnetSpec() {
	ethDataVotes = 2048
	epochAttestations = 4096
	slashings = 8192
	randaoMixes = 65536
	rootsSize = 8192
}
