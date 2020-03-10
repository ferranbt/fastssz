package spectests

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
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
	"github.com/ghodss/yaml"
	baseSSZ "github.com/prysmaticlabs/go-ssz"
)

type codec interface {
	ssz.Marshaler
	ssz.Unmarshaler
}

type testCallback func() codec

var codecs = map[string]testCallback{
	"AggregateAndProof":       func() codec { return new(AggregateAndProof) },
	"Attestation":             func() codec { return new(Attestation) },
	"AttestationData":         func() codec { return new(AttestationData) },
	"AttesterSlashing":        func() codec { return new(AttesterSlashing) },
	"BeaconBlock":             func() codec { return new(BeaconBlock) },
	"BeaconBlockBody":         func() codec { return new(BeaconBlockBody) },
	"BeaconBlockHeader":       func() codec { return new(BeaconBlockHeader) },
	"BeaconState":             func() codec { return new(BeaconState) },
	"Checkpoint":              func() codec { return new(Checkpoint) },
	"Deposit":                 func() codec { return new(Deposit) },
	"DepositData":             func() codec { return new(DepositData) },
	"DepositMessage":          func() codec { return new(DepositMessage) },
	"Eth1Block":               func() codec { return new(Eth1Block) },
	"Eth1Data":                func() codec { return new(Eth1Data) },
	"Fork":                    func() codec { return new(Fork) },
	"HistoricalBatch":         func() codec { return new(HistoricalBatch) },
	"IndexedAttestation":      func() codec { return new(IndexedAttestation) },
	"PendingAttestation":      func() codec { return new(PendingAttestation) },
	"ProposerSlashing":        func() codec { return new(ProposerSlashing) },
	"SignedBeaconBlock":       func() codec { return new(SignedBeaconBlock) },
	"SignedBeaconBlockHeader": func() codec { return new(SignedBeaconBlockHeader) },
	"SignedVoluntaryExit":     func() codec { return new(SignedVoluntaryExit) },
	"SigningRoot":             func() codec { return new(SigningRoot) },
	"Validator":               func() codec { return new(Validator) },
	"VoluntaryExit":           func() codec { return new(VoluntaryExit) },
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
		t.Skip("Fuzz testing not enabled")
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
					t.Fatal("it should have failed")
				}
			}
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
	for _, codec := range codecs {
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

func TestSpecMinimal(t *testing.T) {
	files := readDir(t, filepath.Join(testsPath, "/minimal/phase0/ssz_static"))
	for _, f := range files {
		spl := strings.Split(f, "/")
		name := spl[len(spl)-1]

		base, ok := codecs[name]
		if !ok {
			t.Fatalf("name %s not found", name)
		}

		for _, f := range walkPath(t, f) {
			checkSSZEncoding(t, f, base)
		}
	}
}

func TestSpecMainnet(t *testing.T) {
	files := readDir(t, filepath.Join(testsPath, "/mainnet/phase0/ssz_static"))
	for _, f := range files {
		spl := strings.Split(f, "/")
		name := spl[len(spl)-1]

		if name == "BeaconState" || name == "HistoricalBatch" {
			continue
		}
		base, ok := codecs[name]
		if !ok {
			t.Fatalf("name %s not found", name)
		}

		files := readDir(t, filepath.Join(f, "ssz_random"))
		for _, f := range files {
			checkSSZEncoding(t, f, base)
		}
	}
}

func checkSSZEncoding(t *testing.T, f string, base testCallback) {
	obj := base()
	expected := readValidGenericSSZ(t, f, &obj)

	// Marshal
	res, err := obj.MarshalSSZTo(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(res, expected) {
		t.Fatal("bad")
	}

	// Unmarshal
	obj2 := base()
	if err := obj2.UnmarshalSSZ(expected); err != nil {
		panic(err)
	}
	if !deepEqual(obj, obj2) {
		t.Fatal("bad")
	}
}

const benchmarkTestCase = "../eth2.0-spec-tests/tests/mainnet/phase0/ssz_static/BeaconBlock/ssz_random/case_4"

func BenchmarkMarshalGoSSZ(b *testing.B) {
	obj := new(BeaconBlock)
	readValidGenericSSZ(nil, benchmarkTestCase, obj)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		baseSSZ.Marshal(obj)
	}
}

func BenchmarkMarshalFast(b *testing.B) {
	obj := new(BeaconBlock)
	readValidGenericSSZ(nil, benchmarkTestCase, obj)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		obj.MarshalSSZ()
	}
}

func BenchmarkMarshalSuperFast(b *testing.B) {
	obj := new(BeaconBlock)
	readValidGenericSSZ(nil, benchmarkTestCase, obj)

	buf := make([]byte, 0)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf, _ = obj.MarshalSSZTo(buf[:0])
	}
}

func BenchmarkUnMarshalGoSSZ(b *testing.B) {
	obj := new(BeaconBlock)
	readValidGenericSSZ(nil, benchmarkTestCase, obj)

	dst, err := baseSSZ.Marshal(obj)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var obj2 BeaconBlock
		if err := baseSSZ.Unmarshal(dst, &obj2); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnMarshalFast(b *testing.B) {
	obj := new(BeaconBlock)
	readValidGenericSSZ(nil, benchmarkTestCase, obj)

	dst, err := obj.MarshalSSZ()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		obj2 := new(BeaconBlock)
		if err := obj2.UnmarshalSSZ(dst); err != nil {
			b.Fatal(err)
		}
	}
}

const (
	testsPath      = "../eth2.0-spec-tests/tests"
	serializedFile = "serialized.ssz"
	valueFile      = "value.yaml"
	rootsFile      = "roots.yaml"
)

func walkPath(t *testing.T, path string) (res []string) {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.Contains(path, "case_") {
			res = append(res, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return
}

func readDir(t *testing.T, path string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}
	res := []string{}
	for _, f := range files {
		res = append(res, filepath.Join(path, f.Name()))
	}
	return res
}

func readValidGenericSSZ(t *testing.T, path string, obj interface{}) []byte {
	serialized, err := ioutil.ReadFile(filepath.Join(path, serializedFile))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := ioutil.ReadFile(filepath.Join(path, valueFile))
	if err != nil {
		t.Fatal(err)
	}
	if err := unmarshalYaml(raw, obj); err != nil {
		t.Fatal(err)
	}
	return serialized
}

var hexMatch = regexp.MustCompile("('0[xX][0-9a-fA-F]+')")

func unmarshalYaml(content []byte, obj interface{}) error {
	input := []byte(content)
	for _, match := range hexMatch.FindAllSubmatch(input, -1) {
		res, err := hex.DecodeString(strings.Trim(string(match[1]), "'")[2:])
		if err != nil {
			panic(err)
		}
		resb64 := base64.StdEncoding.EncodeToString(res)
		input = bytes.Replace(input, match[1], []byte(resb64), -1)
	}
	return yaml.Unmarshal(input, &obj)
}
