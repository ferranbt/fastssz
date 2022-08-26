package spectests

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	ssz "github.com/ferranbt/fastssz"
	"github.com/golang/snappy"
	"github.com/prysmaticlabs/gohashtree"

	"gopkg.in/yaml.v2"
)

type codec interface {
	ssz.Marshaler
	ssz.Unmarshaler
	ssz.HashRoot
}

type fork string

const (
	phase0    fork = "phase0"
	altair    fork = "altair"
	bellatrix fork = "bellatrix"
)

type testCallback func(fork fork) codec

var codecs = map[string]testCallback{
	"AttestationData":   func(fork fork) codec { return new(AttestationData) },
	"Checkpoint":        func(fork fork) codec { return new(Checkpoint) },
	"AggregateAndProof": func(fork fork) codec { return new(AggregateAndProof) },
	"Attestation":       func(fork fork) codec { return new(Attestation) },
	"AttesterSlashing":  func(fork fork) codec { return new(AttesterSlashing) },
	"BeaconState": func(fork fork) codec {
		if fork == phase0 {
			return new(BeaconState)
		} else if fork == altair {
			return new(BeaconStateAltair)
		} else if fork == bellatrix {
			return new(BeaconStateBellatrix)
		}
		return nil
	},
	"BeaconBlock": func(fork fork) codec {
		if fork == phase0 {
			return new(BeaconBlock)
		}
		return nil
	},
	"BeaconBlockBody": func(fork fork) codec {
		if fork == phase0 {
			return new(BeaconBlockBodyPhase0)
		} else if fork == altair {
			return new(BeaconBlockBodyAltair)
		} else if fork == bellatrix {
			return new(BeaconBlockBodyBellatrix)
		}
		return nil
	},
	"BeaconBlockHeader":  func(fork fork) codec { return new(BeaconBlockHeader) },
	"Deposit":            func(fork fork) codec { return new(Deposit) },
	"DepositData":        func(fork fork) codec { return new(DepositData) },
	"DepositMessage":     func(fork fork) codec { return new(DepositMessage) },
	"Eth1Block":          func(fork fork) codec { return new(Eth1Block) },
	"Eth1Data":           func(fork fork) codec { return new(Eth1Data) },
	"Fork":               func(fork fork) codec { return new(Fork) },
	"HistoricalBatch":    func(fork fork) codec { return new(HistoricalBatch) },
	"IndexedAttestation": func(fork fork) codec { return new(IndexedAttestation) },
	"PendingAttestation": func(fork fork) codec { return new(PendingAttestation) },
	"ProposerSlashing":   func(fork fork) codec { return new(ProposerSlashing) },
	"SignedBeaconBlock": func(fork fork) codec {
		if fork == phase0 {
			return new(SignedBeaconBlock)
		}
		return nil
	},
	"SignedBeaconBlockHeader": func(fork fork) codec { return new(SignedBeaconBlockHeader) },
	"SignedVoluntaryExit":     func(fork fork) codec { return new(SignedVoluntaryExit) },
	"SigningRoot":             func(fork fork) codec { return new(SigningRoot) },
	"Validator":               func(fork fork) codec { return new(Validator) },
	"VoluntaryExit":           func(fork fork) codec { return new(VoluntaryExit) },
	"ErrorResponse":           func(fork fork) codec { return new(ErrorResponse) },
	"SyncCommittee": func(fork fork) codec {
		return new(SyncCommittee)
	},
	"SyncAggregate": func(fork fork) codec {
		return new(SyncAggregate)
	},
	"ExecutionPayload": func(fork fork) codec {
		return new(ExecutionPayload)
	},
}

func testSpecFork(t *testing.T, fork fork) {
	files := readDir(t, filepath.Join(testsPath, "/mainnet/"+string(fork)+"/ssz_static"))
	for _, f := range files {
		spl := strings.Split(f, "/")
		name := spl[len(spl)-1]

		base, ok := codecs[name]
		if !ok {
			continue
		}

		t.Run(name, func(t *testing.T) {
			files := readDir(t, filepath.Join(f, "ssz_random"))
			for _, f := range files {
				checkSSZEncoding(t, fork, f, name, base)
			}
		})
	}
}

func TestSpec_Phase0(t *testing.T) {
	testSpecFork(t, phase0)
}

func TestSpec_Altair(t *testing.T) {
	testSpecFork(t, altair)
}

func TestSpec_Bellatrix(t *testing.T) {
	testSpecFork(t, bellatrix)
}

func checkSSZEncoding(t *testing.T, fork fork, fileName, structName string, base testCallback) {
	obj := base(fork)
	if obj == nil {
		// skip
		return
	}
	output := readValidGenericSSZ(t, fileName, &obj)

	fatal := func(errHeader string, err error) {
		t.Fatalf("%s spec file=%s, struct=%s, err=%v", errHeader, fileName, structName, err)
	}

	// Marshal
	res, err := obj.MarshalSSZTo(nil)
	if err != nil {
		fatal("marshalSSZ", err)
	}
	if !bytes.Equal(res, output.ssz) {
		fatal("marshalSSZ_equal", fmt.Errorf("bad marshal"))
	}

	// Unmarshal
	obj2 := base(fork)
	if err := obj2.UnmarshalSSZ(output.ssz); err != nil {
		fatal("UnmarshalSSZ", err)
	}
	if !deepEqual(obj, obj2) {
		fatal("UnmarshalSSZ_equal", fmt.Errorf("bad unmarshal"))
	}

	// Root
	root, err := obj.HashTreeRoot()
	if err != nil {
		fatal("HashTreeRoot", err)
	}
	if !bytes.Equal(root[:], output.root) {
		//fmt.Println("- bad root -")
		fatal("HashTreeRoot_equal", fmt.Errorf("bad root"))
	}

	if structName == "BeaconState" || structName == "BeaconBlockBody" || structName == "ExecutionPayload" {
		// this gets to expensive, BeaconState even crashes with out-of-bounds memory allocation
		return
	}

	// Proof
	node, err := obj.GetTree()
	if err != nil {
		fatal("Tree", err)
	}
	nodeRoot := node.Hash()
	if !bytes.Equal(nodeRoot, root[:]) {
		fatal("Tree_equal", fmt.Errorf("bad node"))
	}
}

const benchmarkTestCase = "../eth2.0-spec-tests/tests/mainnet/phase0/ssz_static/BeaconBlock/ssz_random/case_4"

func BenchmarkMarshal_Fast(b *testing.B) {
	obj := new(BeaconBlock)
	readValidGenericSSZ(nil, benchmarkTestCase, obj)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		obj.MarshalSSZ()
	}
}

func BenchmarkMarshal_SuperFast(b *testing.B) {
	obj := new(BeaconBlock)
	readValidGenericSSZ(nil, benchmarkTestCase, obj)

	buf := make([]byte, 0)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf, _ = obj.MarshalSSZTo(buf[:0])
	}
}

func BenchmarkUnMarshal_Fast(b *testing.B) {
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

func BenchmarkHashTreeRoot_Fast(b *testing.B) {
	obj := new(BeaconBlock)
	readValidGenericSSZ(nil, benchmarkTestCase, obj)

	b.ReportAllocs()
	b.ResetTimer()

	hh := ssz.NewHasher()
	for i := 0; i < b.N; i++ {
		obj.HashTreeRootWith(hh)
		hh.Reset()
	}
}

func BenchmarkHashTreeRoot_SuperFast(b *testing.B) {
	obj := new(BeaconBlock)
	readValidGenericSSZ(nil, benchmarkTestCase, obj)

	b.ReportAllocs()
	b.ResetTimer()

	hh := ssz.NewHasherWithHashFn(gohashtree.Hash)
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
	serializedSnappy, err := ioutil.ReadFile(filepath.Join(path, serializedFile))
	if err != nil {
		t.Fatal(err)
	}
	serialized, err := snappy.Decode(nil, serializedSnappy)
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
