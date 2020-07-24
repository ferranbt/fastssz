package spectests

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	ssz "github.com/ferranbt/fastssz"
	"github.com/ferranbt/fastssz/fuzz"
	"github.com/mitchellh/mapstructure"

	"gopkg.in/yaml.v2"
)

type codec interface {
	ssz.Marshaler
	ssz.Unmarshaler
	ssz.HashRoot
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
					t.Fatalf("%s it should have failed", name)
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

		t.Log(f)
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

		t.Log(f)
		files := readDir(t, filepath.Join(f, "ssz_random"))
		for _, f := range files {
			checkSSZEncoding(t, f, base)
		}
	}
}

func checkSSZEncoding(t *testing.T, f string, base testCallback) {
	obj := base()
	output := readValidGenericSSZ(t, f, &obj)

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
		panic(err)
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
		fmt.Printf("%s bad root\n", f)
	}
}

const benchmarkTestCase = "../eth2.0-spec-tests/tests/mainnet/phase0/ssz_static/BeaconBlock/ssz_random/case_4"

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

func BenchmarkHashTreeRootFast(b *testing.B) {
	obj := new(BeaconBlock)
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

	if err := UnmarshalSSZTest(raw, obj); err != nil {
		t.Fatal(err)
	}
	return &output{root: root, ssz: serialized}
}

func isByteSlice(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8
}

func isByteArray(t reflect.Type) bool {
	return t.Kind() == reflect.Array && t.Elem().Kind() == reflect.Uint8
}

func customHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}

	raw := data.(string)
	if !strings.HasPrefix(raw, "0x") {
		return nil, fmt.Errorf("0x prefix not found")
	}
	elem, err := hex.DecodeString(raw[2:])
	if err != nil {
		return nil, err
	}
	if isByteSlice(t) {
		// []byte
		return elem, nil
	}
	if isByteArray(t) {
		// [n]byte
		if t.Len() != len(elem) {
			return nil, fmt.Errorf("incorrect array length: %d %d", t.Len(), len(elem))
		}

		v := reflect.New(t)
		reflect.Copy(v.Elem(), reflect.ValueOf(elem))
		return v.Interface(), nil
	}

	return nil, fmt.Errorf("type not found")
}

func UnmarshalSSZTest(content []byte, result interface{}) error {
	var source map[string]interface{}
	if err := yaml.Unmarshal(content, &source); err != nil {
		return err
	}

	dc := &mapstructure.DecoderConfig{
		Result:     result,
		DecodeHook: customHook,
		TagName:    "json",
	}
	ms, err := mapstructure.NewDecoder(dc)
	if err != nil {
		return err
	}
	if err = ms.Decode(source); err != nil {
		return err
	}
	return nil
}
