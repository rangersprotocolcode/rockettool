package bls

import (
	bn_curve "RocketTool/src/ecdsa/bls/bn256"
	"RocketTool/src/util"
	"fmt"
	"log"
)

type Pubkey struct {
	value bn_curve.G2
}

func GeneratePubkey(sec Seckey) *Pubkey {
	pub := new(Pubkey)
	pub.value.ScalarBaseMult(sec.value.getBigInt())
	return pub
}

func (pub Pubkey) Serialize() []byte {
	return pub.value.Marshal()
}

func (pub *Pubkey) Deserialize(b []byte) error {
	_, error := pub.value.Unmarshal(b)
	return error
}

func (pub Pubkey) GetHexString() string {
	return PREFIX + util.Bytes2Hex(pub.value.Marshal())
}

func (pub *Pubkey) SetHexString(s string) error {
	if len(s) < len(PREFIX) || s[:len(PREFIX)] != PREFIX {
		return fmt.Errorf("arg failed")
	}
	buf := s[len(PREFIX):]

	pub.value.Unmarshal(util.Hex2Bytes(buf))
	return nil
}

func (pub Pubkey) IsEmpty() bool {
	return pub.value.IsEmpty()
}

func (pub Pubkey) IsValid() bool {
	return !pub.IsEmpty()
}

func AggregatePubkeys(pubs []Pubkey) *Pubkey {
	if len(pubs) == 0 {
		log.Printf("AggregatePubkeys no pubs")
		return nil
	}

	pub := new(Pubkey)
	pub.value.Set(&pubs[0].value)

	for i := 1; i < len(pubs); i++ {
		pub.add(&pubs[i])
	}

	return pub
}

func (pub *Pubkey) add(rhs *Pubkey) error {
	pa := &bn_curve.G2{}
	pb := &bn_curve.G2{}

	pa.Set(&pub.value)
	pb.Set(&rhs.value)

	pub.value.Add(pa, pb)
	return nil
}

func (pub Pubkey) MarshalJSON() ([]byte, error) {
	str := "\"" + pub.GetHexString() + "\""
	return []byte(str), nil
}

func (pub *Pubkey) UnmarshalJSON(data []byte) error {
	str := string(data[:])
	if len(str) < 2 {
		return fmt.Errorf("data size less than min.")
	}
	str = str[1 : len(str)-1]
	return pub.SetHexString(str)
}
