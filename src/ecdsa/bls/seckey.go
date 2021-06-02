package bls

import (
	"log"
	"math/big"

	bn_curve "RocketTool/src/ecdsa/bls/bn256"
	"RocketTool/src/util"
)

// Curve and Field order
var curveOrder = bn_curve.Order
var fieldOrder = bn_curve.P
var bitLength = curveOrder.BitLen()

type Seckey struct {
	value BnInt
}

func NewSeckeyFromRand(seed util.Rand) *Seckey {
	return newSeckeyFromByte(seed.Bytes())
}

func NewSeckeyFromBigInt(b *big.Int) *Seckey {
	nb := &big.Int{}
	nb.Set(b)
	b.Mod(nb, curveOrder)

	sec := new(Seckey)
	sec.value.setBigInt(b)

	return sec
}

func (sec Seckey) Serialize() []byte {
	return sec.value.serialize()
}

func (sec *Seckey) Deserialize(b []byte) error {
	return sec.value.deserialize(b)
}

func (sec Seckey) GetBigInt() (s *big.Int) {
	s = new(big.Int)
	s.Set(sec.value.getBigInt())
	return s
}

func (sec Seckey) GetHexString() string {
	return sec.value.getHexString()
}

func AggregateSeckeys(secs []Seckey) *Seckey {
	if len(secs) == 0 {
		log.Printf("AggregateSeckeys no secs")
		return nil
	}
	sec := new(Seckey)
	sec.value.setBigInt(secs[0].value.getBigInt())
	for i := 1; i < len(secs); i++ {
		sec.value.add(&secs[i].value)
	}

	x := new(big.Int)
	x.Set(sec.value.getBigInt())
	sec.value.setBigInt(x.Mod(x, curveOrder))
	return sec
}

func ShareSeckey(msec []Seckey, id ID) *Seckey {
	secret := big.NewInt(0)
	k := len(msec) - 1

	secret.Set(msec[k].GetBigInt())
	x := id.GetBigInt()
	new_b := &big.Int{}

	for j := k - 1; j >= 0; j-- {
		new_b.Set(secret)
		secret.Mul(new_b, x)

		new_b.Set(secret)
		secret.Add(new_b, msec[j].GetBigInt())

		new_b.Set(secret)
		secret.Mod(new_b, curveOrder)
	}

	return NewSeckeyFromBigInt(secret)
}

func newSeckeyFromByte(b []byte) *Seckey {
	sec := new(Seckey)
	err := sec.Deserialize(b[:32])
	if err != nil {
		log.Printf("NewSeckeyFromByte %s\n", err)
		return nil
	}

	sec.value.mod()
	return sec
}
