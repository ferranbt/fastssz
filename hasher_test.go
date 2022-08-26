package ssz

import (
	"fmt"
	"testing"

	"github.com/prysmaticlabs/gohashtree"
)

func TestDepth(t *testing.T) {
	cases := []struct {
		Num uint64
		Res uint8
	}{
		{0, 0},
		{1, 0},
		{2, 1},
		{3, 2},
		{4, 2},
		{5, 3},
		{6, 3},
		{7, 3},
		{8, 3},
		{9, 4},
		{16, 4},
		{1024, 10},
	}
	for _, c := range cases {
		if depth := getDepth(c.Num); depth != c.Res {
			t.Fatalf("num %d, expected %d but found %d", c.Num, c.Res, depth)
		}
	}
}

func TestNextPowerOfTwo(t *testing.T) {
	cases := []struct {
		Num, Res uint64
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{6, 8},
		{7, 8},
		{8, 8},
		{9, 16},
		{10, 16},
		{11, 16},
		{13, 16},
	}
	for _, c := range cases {
		if next := nextPowerOfTwo(c.Num); uint64(next) != c.Res {
			t.Fatalf("num %d, expected %d but found %d", c.Num, c.Res, next)
		}
	}
}

func TestHashGoHashTree(t *testing.T) {

	a := make([]byte, 32)
	b := make([]byte, 32)
	a[0] = 1
	b[0] = 2

	buf := []byte{}
	buf = append(buf, a...)
	buf = append(buf, b...)

	gohashtree.Hash(buf, buf)

	fmt.Println(buf)
}
