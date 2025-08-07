package spectests

var (
	ethDataVotes      int
	epochAttestations int
	slashings         int
	randaoMixes       int
	rootsSize         int
)

func init() {
	setMainnetSpec()
}

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
