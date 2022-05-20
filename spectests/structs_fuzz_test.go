package spectests

import (
	"bytes"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/ferranbt/fastssz/fuzz"
)

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
			obj := codec("")

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
	codec := codecs["ErrorResponse"]

	for i := 0; i < 1000; i++ {
		obj := codec("")
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

		obj2 := codec("")
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
			obj := codec("")
			f := fuzz.New()
			f.Fuzz(obj)

			dst, err := obj.MarshalSSZTo(nil)
			if err != nil {
				t.Fatal(err)
			}

			obj2 := codec("")
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
			obj := codec("")
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

				obj2 := codec("")
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
		obj := codec("")
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
			obj2 := codec("")
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
