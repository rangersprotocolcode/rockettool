package bls

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"

	bn_curve "RocketTool/src/ecdsa/bls/bn256"
	"RocketTool/src/util"
)

type Signature struct {
	value bn_curve.G1
}

func DeserializeSign(b []byte) *Signature {
	sig := &Signature{}
	sig.Deserialize(b)
	return sig
}

func (sig Signature) Serialize() []byte {
	if sig.IsNil() {
		return []byte{}
	}
	return sig.value.Marshal()
}

func (sig *Signature) Deserialize(b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("signature Deserialized failed.")
	}
	sig.value.Unmarshal(b)
	return nil
}

func (sig Signature) GetHexString() string {
	return PREFIX + util.Bytes2Hex(sig.value.Marshal())
}

func (sig *Signature) IsNil() bool {
	return sig.value.IsNil()
}

func (sig Signature) IsEqual(rhs Signature) bool {
	return bytes.Equal(sig.value.Marshal(), rhs.value.Marshal())
}

func (sig Signature) IsValid() bool {
	s := sig.Serialize()
	if len(s) == 0 {
		return false
	}

	return sig.value.IsValid()
}

func Sign(sec Seckey, msg []byte) (sig Signature) {
	bg := hashToG1(string(msg))
	sig.value.ScalarMult(bg, sec.GetBigInt())
	return sig
}

func VerifySig(pub Pubkey, msg []byte, sig Signature) bool {
	if sig.IsNil() || !sig.IsValid() {
		return false
	}
	if !pub.IsValid() {
		return false
	}
	if sig.value.IsNil() {
		return false
	}
	bQ := bn_curve.GetG2Base()
	p1 := bn_curve.Pair(&sig.value, bQ)

	Hm := hashToG1(string(msg))
	p2 := bn_curve.Pair(Hm, &pub.value)

	return bn_curve.PairIsEuqal(p1, p2)
}

func RecoverGroupSignature(memberSignMap map[string]Signature, thresholdValue int) *Signature {
	if thresholdValue < len(memberSignMap) {
		memberSignMap = getRandomKSignInfo(memberSignMap, thresholdValue)
	}
	ids := make([]ID, thresholdValue)
	sigs := make([]Signature, thresholdValue)
	i := 0
	for s_id, si := range memberSignMap {
		var id ID
		id.SetHexString(s_id)
		ids[i] = id
		sigs[i] = si
		i++
		if i >= thresholdValue {
			break
		}
	}
	return recoverSignature(sigs, ids)
}

func getRandomKSignInfo(memberSignMap map[string]Signature, k int) map[string]Signature {
	indexs := util.NewRand().RandomPerm(len(memberSignMap), k)
	sort.Ints(indexs)
	ret := make(map[string]Signature)

	i := 0
	j := 0
	for key, sign := range memberSignMap {
		if i == indexs[j] {
			ret[key] = sign
			j++
			if j >= k {
				break
			}
		}
		i++
	}
	return ret
}

func recoverSignature(sigs []Signature, ids []ID) *Signature {
	k := len(sigs)
	xs := make([]*big.Int, len(ids))
	for i := 0; i < len(xs); i++ {
		xs[i] = ids[i].GetBigInt()
	}
	sig := &Signature{}
	new_sig := &Signature{}
	for i := 0; i < k; i++ {
		var delta, num, den, diff *big.Int = big.NewInt(1), big.NewInt(1), big.NewInt(1), big.NewInt(0)
		for j := 0; j < k; j++ {
			if j != i {
				num.Mul(num, xs[j])
				num.Mod(num, curveOrder)
				diff.Sub(xs[j], xs[i])
				den.Mul(den, diff)
				den.Mod(den, curveOrder)
			}
		}
		den.ModInverse(den, curveOrder)
		delta.Mul(num, den)
		delta.Mod(delta, curveOrder)

		new_sig.value.Set(&sigs[i].value)
		new_sig.mul(delta)

		if i == 0 {
			sig.value.Set(&new_sig.value)
		} else {
			sig.add(new_sig)
		}
	}
	return sig
}

func (sig *Signature) add(sig1 *Signature) error {
	new_sig := &Signature{}
	new_sig.value.Set(&sig.value)
	sig.value.Add(&new_sig.value, &sig1.value)

	return nil
}

func (sig *Signature) mul(bi *big.Int) error {
	g1 := new(bn_curve.G1)
	g1.Set(&sig.value)
	sig.value.ScalarMult(g1, bi)
	return nil
}
