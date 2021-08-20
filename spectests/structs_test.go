package spectests

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ferranbt/fastssz/spectests/phase0"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	ssz "github.com/ferranbt/fastssz"
	"github.com/ferranbt/fastssz/fuzz"

	"github.com/golang/snappy"
	"gopkg.in/yaml.v2"
)

type codec interface {
	ssz.Marshaler
	ssz.Unmarshaler
	ssz.HashRoot
}

type codecTree interface {
	GetTreeWithWrapper(w *ssz.Wrapper) (err error)
	GetTree() (*ssz.Node, error)
}

type testCallback func() codec

func valueForContainerPhase0(name string) (codec, error) {
	switch name {
	case "AggregateAndProof":
		return &phase0.AggregateAndProof{}, nil
	case "Attestation":
		return &phase0.Attestation{}, nil
	case "AttestationData":
		return &phase0.AttestationData{}, nil
	case "AttesterSlashing":
		return &phase0.AttesterSlashing{}, nil
	case "BeaconBlock":
		return &phase0.BeaconBlock{}, nil
	case "BeaconBlockBody":
		return &phase0.BeaconBlockBody{}, nil
	case "BeaconBlockHeader":
		return &phase0.BeaconBlockHeader{}, nil
	case "BeaconState":
		return &phase0.BeaconState{}, nil
	case "Checkpoint":
		return &phase0.Checkpoint{}, nil
	case "Deposit":
		return &phase0.Deposit{}, nil
	case "DepositData":
		return &phase0.DepositData{}, nil
	case "DepositMessage":
		return &phase0.DepositMessage{}, nil
	case "Eth1Block":
		return &phase0.Eth1Block{}, nil
	case "Eth1Data":
		return &phase0.Eth1Data{}, nil
	case "Fork":
		return &phase0.Fork{}, nil
	case "HistoricalBatch":
		return &phase0.HistoricalBatch{}, nil
	case "IndexedAttestation":
		return &phase0.IndexedAttestation{}, nil
	case "PendingAttestation":
		return &phase0.PendingAttestation{}, nil
	case "ProposerSlashing":
		return &phase0.ProposerSlashing{}, nil
	case "SignedBeaconBlock":
		return &phase0.SignedBeaconBlock{}, nil
	case "SignedBeaconBlockHeader":
		return &phase0.SignedBeaconBlockHeader{}, nil
	case "SignedVoluntaryExit":
		return &phase0.SignedVoluntaryExit{}, nil
	case "SigningRoot":
		return &phase0.SigningRoot{}, nil
	case "Validator":
		return &phase0.Validator{}, nil
	case "VoluntaryExit":
		return &phase0.VoluntaryExit{}, nil
	case "ErrorResponse":
		return &phase0.ErrorResponse{}, nil
	default:
		return nil, fmt.Errorf("unknown container named %s", name)
	}
}

var phase0Containers = []string{"AggregateAndProof","Attestation","AttestationData","AttesterSlashing","BeaconBlock",
	"BeaconBlockBody","BeaconBlockHeader","BeaconState","Checkpoint","Deposit","DepositData","DepositMessage",
	"Eth1Block","Eth1Data","Fork","HistoricalBatch","IndexedAttestation","PendingAttestation","ProposerSlashing",
	"SignedBeaconBlock","SignedBeaconBlockHeader","SignedVoluntaryExit","SigningRoot","Validator","VoluntaryExit",
	"ErrorResponse"}

func valueForContainerAltair(name string) (codec, error) {
	switch name {
	default:
		return valueForContainerPhase0(name)
	}
}

func valueForContainer(fork string, container string) (codec, error) {
	switch fork {
	case "phase0":
		return valueForContainerPhase0(container)
	case "altair":
		return valueForContainerAltair(container)
	// TODO: only BeaconBlockBody and BeaconState changed, but BeaconBlockBody is included by
	// several other containers, so any containers that use it also need to be duplicated into
	// the altair package so that they refer to the right version. otherwise in altair, any
	// dependencies that are unchanged should point to the phase0 version.
	default:
		return nil, fmt.Errorf("spectests do not know about fork named %s", fork)
	}
}

var altairContainers = []string{"SyncCommitteeDuty"}

var codecs = map[string]testCallback{
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

const defaultFuzzCount = 100

func fuzzTestCount(t *testing.T, i string) int {
	var num int

	// read the number of tests from an env variable
	numStr := os.Getenv("FUZZ_COUNT")
	if numStr == "" {
		num = defaultFuzzCount
	} else {
		var err error
		if num, err = strconv.Atoi(numStr); err != nil {
			t.Fatal(err)
		}
	}

	if i == "BeaconState" {
		// BeaconState is too big to run all the cases, execute
		// only a 10% of them
		return int(.1 * float64(num))
	}
	return num
}

func checkIsFuzzEnabled(t *testing.T) {
	if strings.ToLower(os.Getenv("FUZZ_TESTS")) != "true" {
		t.Skip("Fuzz testing not enabled, skipping")
	}
}

func TestFuzzMarshalWithWrongSizes(t *testing.T) {
	checkIsFuzzEnabled(t)

	for name, codec := range codecs {
		count := fuzzTestCount(t, name)
		for i := 0; i < count; i++ {
			obj := codec()

			f := fuzz.New()
			f.SetFailureRatio(.1)

			failed := f.Fuzz(obj)
			if failed {
				if _, err := obj.MarshalSSZTo(nil); err == nil {
					t.Fatalf("%s it should have failed", name)
				}
			}
		}
	}
}

func TestErrorResponse(t *testing.T) {
	// TODO: Move to fuzzer
	codec := codecs["ErrorResponse"]

	for i := 0; i < 1000; i++ {
		obj := codec()
		f := fuzz.New()
		f.SetFailureRatio(.1)
		failed := f.Fuzz(obj)

		dst, err := obj.MarshalSSZTo(nil)
		if err != nil {
			if !failed {
				t.Fatal(err)
			} else {
				continue
			}
		}

		obj2 := codec()
		if err := obj2.UnmarshalSSZ(dst); err != nil {
			t.Fatal(err)
		}
		if !deepEqual(obj, obj2) {
			t.Fatal("bad")
		}
	}
}

func TestFuzzEncoding(t *testing.T) {
	checkIsFuzzEnabled(t)

	for name, codec := range codecs {
		count := fuzzTestCount(t, name)
		for i := 0; i < count; i++ {
			obj := codec()
			f := fuzz.New()
			f.Fuzz(obj)

			dst, err := obj.MarshalSSZTo(nil)
			if err != nil {
				t.Fatal(err)
			}

			obj2 := codec()
			if err := obj2.UnmarshalSSZ(dst); err != nil {
				t.Fatal(err)
			}
			if !deepEqual(obj, obj2) {
				t.Fatal("bad")
			}
		}
	}
}

func TestFuzzUnmarshalAppend(t *testing.T) {
	checkIsFuzzEnabled(t)

	// Fuzz with append values between the fields
	for name, codec := range codecs {
		t.Logf("Process %s", name)

		for j := 0; j < 5; j++ {
			obj := codec()
			f := fuzz.New()
			f.Fuzz(obj)

			dst, err := obj.MarshalSSZTo(nil)
			if err != nil {
				t.Fatal(err)
			}

			for i := 0; i < 100; i++ {
				buf := []byte{}

				pos := randomInt(0, len(dst))
				size := randomInt(1, 20)

				aux := make([]byte, size)
				rand.Read(aux)

				buf = append(buf, dst[:pos]...)
				buf = append(buf, aux...)
				buf = append(buf, dst[pos:]...)

				obj2 := codec()
				if err := obj2.UnmarshalSSZ(buf); err == nil {
					if deepEqual(obj, obj2) {
						t.Fatal("bad")
					}
				}
			}
		}
	}
}

func TestFuzzUnmarshalShuffle(t *testing.T) {
	checkIsFuzzEnabled(t)

	// Unmarshal a correct dst with shuffled data
	for _, codec := range codecs {
		obj := codec()
		f := fuzz.New()
		f.Fuzz(obj)

		dst, err := obj.MarshalSSZTo(nil)
		if err != nil {
			t.Fatal(err)
		}

		for i := 0; i < 100; i++ {
			buf := make([]byte, len(dst))
			copy(buf, dst)

			pos := randomInt(1, len(dst))
			n := randomInt(2, 4)
			rand.Read(buf[pos:min(pos+n, len(dst))])

			if bytes.Equal(buf, dst) {
				continue
			}
			obj2 := codec()
			if err := obj2.UnmarshalSSZ(buf); err == nil {
				if deepEqual(obj, obj2) {
					t.Fatal("bad")
				}
			}
		}
	}

}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

// example: eth2.0-spec-tests/tests/mainnet/phase0/ssz_static/Attestation/ssz_random/case_0/
type specTestCase struct {
	testSet string // ex: mainnet
	fork string // ex: phase0
	containerName string // ex: Attestation
	caseNumber string // ex: case_0
	path string // path to test case directory
	root []byte // expected hash tree root of container, read from roots.yaml
	serializedValue []byte // decompressed, serialized value for container fixture, read from serialized.ssz_snappy
	yamlValue []byte // yaml representation of the value, used by fastssz's unmarshal test
	skipContainers map[string]bool // list of container names to skip
}

func newSpecTestCase(path string, testSet string, fork string, containerName string, caseNumber string) *specTestCase {
	return &specTestCase{
		testSet:       testSet,
		fork:          fork,
		containerName: containerName,
		caseNumber:    caseNumber,
		path:          path,
	}
}

func (tc *specTestCase) prepare() error {
	// read snappy-encoded, ssz serialized bytes for test case from serialized.ssz_snappy
	snappyBytes, err := os.ReadFile(filepath.Join(tc.path, serializedFile))
	if err != nil {
		return fmt.Errorf("failed to read snappy bytes for case=%s with err=%s", tc.Name(), err)
	}
	serializedValue, err := snappy.Decode(nil, snappyBytes)
	if err != nil {
		return fmt.Errorf("failed to read snappy bytes serialized ssz for case=%s with err=%s", tc.Name(), err)
	}
	tc.serializedValue = serializedValue

	// read expected hash tree root from roots.yaml
	rootBytes, err := os.ReadFile(filepath.Join(tc.path, rootsFile))
	if err != nil {
		return fmt.Errorf("failed to read %s for case=%s with err=%s", rootsFile, tc.Name(), err)
	}
	rootStruct := struct{Root string `json:"root"`}{}
	if err := yaml.Unmarshal(rootBytes, &rootStruct); err != nil {
		return fmt.Errorf("failed to decode yaml root key from %s for case=%s with err=%s", rootsFile, tc.Name(), err)
	}
	root, err := hex.DecodeString(rootStruct.Root[2:])
	if err != nil {
		return fmt.Errorf("failed to hex decode root key from %s for case=%s with err=%s", rootsFile, tc.Name(), err)
	}
	tc.root = root

	// read the yaml-encoded version of the value
	yamlBytes, err := os.ReadFile(filepath.Join(tc.path, valueFile))
	if err != nil {
		return fmt.Errorf("failed to read %s for case=%s with err=%s", valueFile, tc.Name(), err)
	}
	tc.yamlValue = yamlBytes

	return nil
}

func (tc specTestCase) Runner(t *testing.T) {
	if _, skip := tc.skipContainers[tc.containerName]; skip {
		t.Skip("container type skipped in this config+fork")
	}
	err := tc.prepare()
	if err != nil {
		t.Fatal(err)
	}
	v, err := valueForContainer(tc.fork, tc.containerName)
	if err != nil {
		t.Fatalf("no type definition for type=%s", tc.containerName)
	}
	err = v.UnmarshalSSZ(tc.serializedValue)
	if err != nil {
		t.Fatalf("UnmarshalSSZ failure: %s", err)
	}
	serialized, err := v.MarshalSSZ()
	if err != nil {
		t.Fatalf("MarshalSSZ failure: %s", err)
	}
	if !bytes.Equal(serialized, tc.serializedValue) {
		t.Errorf("MarshalSSZ produced a different value from the serialized ssz fixture")
	}
	vfresh, _ := valueForContainer(tc.fork, tc.containerName)
	err = vfresh.UnmarshalSSZ(serialized)
	if err != nil {
		t.Fatalf("using serialized value produced by MarshalSSZ, UnmarshalSSZ failed with err: %s", err)
	}
	htr, err := vfresh.HashTreeRoot()
	if err != nil {
		t.Fatalf("error from calling HashTreeRoot: %s", err)
	}
	if !bytes.Equal(htr[:], tc.root) {
		t.Fatalf("result of calling HashTreeRoot does not match the expected value")
	}
}

func (tc specTestCase) Name() string {
	//return fmt.Sprintf("%s-%s-%s-%s", tc.testSet, tc.fork, tc.containerName, tc.caseNumber)
	return tc.path
}

// example: eth2.0-spec-tests/tests/mainnet/phase0/ssz_static/Attestation/ssz_random/case_0/
var testcaseRE = regexp.MustCompile(`.*\/ssz_static\/(.+)\/ssz_random\/([^\/]+)`)

// Run all spec tests for mainnet and minimal configurations, across phase0 and altair
// This test is a meta test that creates a separate subtest for each leaf case in the spectests tree
func TestSpectests(t *testing.T) {
	parentCases := []specTestCase{
		/*
		{
			testSet: "minimal",
			fork: "phase0",
		},
		{
			testSet: "minimal",
			fork: "altair",
			skipContainers: map[string]bool{"BeaconState": true, "BeaconBlockBody": true},
		},
		 */
		{
			testSet: "mainnet",
			fork: "phase0",
			skipContainers: map[string]bool{"BeaconState": true, "HistoricalBatch": true},
		},
		/*
		{
			testSet: "mainnet",
			fork: "altair",
			skipContainers: map[string]bool{"BeaconState": true, "BeaconBlockBody": true},
		},
		 */
	}
	for _, c := range parentCases {
		parentPath := filepath.Join(testsPath, c.testSet, c.fork, "ssz_static")
		err := filepath.WalkDir(parentPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				t.Logf("traversing directory %s, failed with error '%s' - skipping this tree of tests", path, err)
				return err
			}
			// skip individual files and parent directories
			if !d.IsDir() || !testcaseRE.MatchString(path){
				return nil
			}
			parts := testcaseRE.FindStringSubmatch(path)
			// the following nested array indexing is safe because the MatchString above
			// guarantees the string structure will match 2 regexp groups
			tc := newSpecTestCase(path, c.testSet, c.fork, parts[1], parts[2])
			t.Run(tc.Name(), tc.Runner)
			return nil
		})
		if err != nil {
			t.Logf("reading spectests from directory %s, failed with error '%s' - skipping this tree of tests", parentPath, err)
			continue
		}
	}
}

func formatSpecFailure(errHeader, specFile, structName string, err error) string {
	return fmt.Sprintf("%s spec file=%s, struct=%s, err=%v",
		errHeader, specFile, structName, err)
}

func checkSSZEncoding(t *testing.T, fileName, structName string, base testCallback) {
	obj := base()
	output := readValidGenericSSZ(t, fileName, &obj)

	// Marshal
	res, err := obj.MarshalSSZTo(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(res, output.ssz) {
		t.Fatal("bad marshalling")
	}

	// Unmarshal
	obj2 := base()
	if err := obj2.UnmarshalSSZ(res); err != nil {
		t.Fatal(formatSpecFailure("UnmarshalSSZ error", fileName, structName, err))
	}
	if !deepEqual(obj, obj2) {
		t.Fatal("bad unmarshalling")
	}

	// Root
	root, err := obj.HashTreeRoot()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(root[:], output.root) {
		fmt.Printf("%s bad root\n", fileName)
	}

	if objt, ok := obj.(codecTree); ok {
		// node root
		node, err := objt.GetTree()
		if err != nil {
			t.Fatal(err)
		}

		xx := node.Hash()
		if !bytes.Equal(xx, root[:]) {
			t.Fatal("bad node")
		}
	}
}

const benchmarkTestCase = "../eth2.0-spec-tests/tests/mainnet/phase0/ssz_static/BeaconBlock/ssz_random/case_4"

func BenchmarkMarshalFast(b *testing.B) {
	obj := new(phase0.BeaconBlock)
	readValidGenericSSZ(nil, benchmarkTestCase, obj)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		obj.MarshalSSZ()
	}
}

func BenchmarkMarshalSuperFast(b *testing.B) {
	obj := new(phase0.BeaconBlock)
	readValidGenericSSZ(nil, benchmarkTestCase, obj)

	buf := make([]byte, 0)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf, _ = obj.MarshalSSZTo(buf[:0])
	}
}

func BenchmarkUnMarshalFast(b *testing.B) {
	obj := new(phase0.BeaconBlock)
	readValidGenericSSZ(nil, benchmarkTestCase, obj)

	dst, err := obj.MarshalSSZ()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		obj2 := new(phase0.BeaconBlock)
		if err := obj2.UnmarshalSSZ(dst); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHashTreeRootFast(b *testing.B) {
	obj := new(phase0.BeaconBlock)
	readValidGenericSSZ(nil, benchmarkTestCase, obj)

	b.ReportAllocs()
	b.ResetTimer()

	hh := ssz.DefaultHasherPool.Get()
	for i := 0; i < b.N; i++ {
		obj.HashTreeRootWith(hh)
		hh.Reset()
	}
}

const (
	testsPath      = "../eth2.0-spec-tests/tests"
	serializedFile = "serialized.ssz_snappy"
	valueFile      = "value.yaml"
	rootsFile      = "roots.yaml"
)

type output struct {
	root []byte
	ssz  []byte
}

func readValidGenericSSZ(t *testing.T, path string, obj interface{}) *output {
	serialized, err := ioutil.ReadFile(filepath.Join(path, serializedFile))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := ioutil.ReadFile(filepath.Join(path, valueFile))
	if err != nil {
		t.Fatal(err)
	}
	raw2, err := ioutil.ReadFile(filepath.Join(path, rootsFile))
	if err != nil {
		t.Fatal(err)
	}

	// Decode ssz root
	var out map[string]string
	if err := yaml.Unmarshal(raw2, &out); err != nil {
		t.Fatal(err)
	}
	root, err := hex.DecodeString(out["root"][2:])
	if err != nil {
		t.Fatal(err)
	}

	if err := ssz.UnmarshalSSZTest(raw, obj); err != nil {
		t.Fatal(err)
	}
	return &output{root: root, ssz: serialized}
}
