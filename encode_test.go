package ssz

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestBitlist(t *testing.T) {
	res := []string{}
	err := filepath.Walk("./spectests/fixtures/bitlist", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			res = append(res, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, i := range res {
		t.Run(i, func(t *testing.T) {
			serialized, err := ioutil.ReadFile(i)
			if err != nil {
				t.Fatal(err)
			}

			var maxSize uint64

			testName := strings.Split(i, "/")[3]
			if strings.Contains(testName, "_no_") {
				maxSize = 1000
			} else {
				// decode maxSize from name
				num, err := strconv.Atoi(strings.Split(testName, "_")[1])
				if err != nil {
					t.Fatal(err)
				}
				maxSize = uint64(num)
			}

			if err := ValidateBitlist(serialized, maxSize); err == nil {
				t.Fatal("bad")
			}
		})
	}
}
