package spectests

import (
	"fmt"
	"math/big"
)

func (u *Uint256) UnmarshalText(text []byte) error {
	x := new(big.Int)
	if err := x.UnmarshalText(text); err != nil {
		return err
	}
	if x.BitLen() > 256 {
		return fmt.Errorf("too big")
	}

	buf := reverse(x.Bytes())
	copy(u[:], buf)
	return nil
}

func (u Uint256) MarshalText() (text []byte, err error) {
	buf := reverse(u[:])
	x := new(big.Int).SetBytes(buf[:])
	return []byte(x.String()), nil
}

func reverse(in []byte) (out []byte) {
	out = make([]byte, len(in))
	copy(out[:], in[:])

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return
}
