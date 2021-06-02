package bls

import (
	"math/big"

	bn_curve "RocketTool/src/ecdsa/bls/bn256"
	"fmt"
)

const PREFIX = "0x"

func hashToG1(m string) *bn_curve.G1 {
	g := &bn_curve.G1{}
	g.HashToPoint([]byte(m))
	return g
}

type BnInt struct {
	v big.Int
}

func (bi *BnInt) getHexString() string {
	buf := bi.v.Text(16)
	return PREFIX + buf
}

func (bi *BnInt) setHexString(s string) error {
	if len(s) < len(PREFIX) || s[:len(PREFIX)] != PREFIX {
		return fmt.Errorf("arg failed")
	}
	buf := s[len(PREFIX):]
	bi.v.SetString(buf[:], 16)
	return nil
}

func (bi *BnInt) getBigInt() *big.Int {
	return new(big.Int).Set(&bi.v)
}

func (bi *BnInt) setBigInt(b *big.Int) error {
	bi.v.Set(b)
	return nil
}

func (bi *BnInt) serialize() []byte {
	return bi.v.Bytes()
}

func (bi *BnInt) deserialize(b []byte) error {
	bi.v.SetBytes(b)
	return nil
}

func (bi *BnInt) isEqual(b *BnInt) bool {
	return 0 == bi.v.Cmp(&b.v)
}

func (bi *BnInt) add(b *BnInt) error {
	bi.v.Add(&bi.v, &b.v)
	return nil
}

func (bi *BnInt) sub(b *BnInt) error {
	bi.v.Sub(&bi.v, &b.v)
	return nil
}

func (bi *BnInt) mul(b *BnInt) error {
	bi.v.Mul(&bi.v, &b.v)
	return nil
}

func (bi *BnInt) mod() error {
	bi.v.Mod(&bi.v, bn_curve.Order)
	return nil
}
