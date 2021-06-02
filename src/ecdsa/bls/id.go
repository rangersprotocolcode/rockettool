package bls

import (
	"math/big"

	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/sha3"
	"log"
)

const ID_LENGTH = 32

type ID struct {
	value BnInt
}

func NewIDFromPubkey(pk Pubkey) *ID {
	h := sha3.Sum256(pk.Serialize())
	bi := new(big.Int).SetBytes(h[:])
	return newIDFromBigInt(bi)
}

func (id ID) GetHexString() string {
	bs := id.Serialize()

	hex := hex.EncodeToString(bs)
	if len(hex) == 0 {
		hex = "0"
	}
	return "0x" + hex
}

func (id ID) GetBigInt() *big.Int {
	x := new(big.Int)
	x.Set(id.value.getBigInt())
	return x
}

func (id *ID) SetHexString(s string) error {
	return id.value.setHexString(s)
}

//把字节切片转换到ID
func (id *ID) Deserialize(b []byte) error {
	return id.value.deserialize(b)
}

func (id ID) Serialize() []byte {
	idBytes := id.value.serialize()
	if len(idBytes) == ID_LENGTH {
		return idBytes
	}
	if len(idBytes) > ID_LENGTH {
		panic("ID Serialize error: ID bytes is more than IDLENGTH")
	}
	buff := make([]byte, ID_LENGTH)
	copy(buff[ID_LENGTH-len(idBytes):ID_LENGTH], idBytes)
	return buff
}

func (id ID) MarshalJSON() ([]byte, error) {
	str := "\"" + id.GetHexString() + "\""
	return []byte(str), nil
}

func (id *ID) UnmarshalJSON(data []byte) error {
	str := string(data[:])
	if len(str) < 2 {
		return fmt.Errorf("data size less than min.")
	}
	str = str[1 : len(str)-1]
	return id.SetHexString(str)
}

func newIDFromBigInt(b *big.Int) *ID {
	id := new(ID)
	err := id.value.setBigInt(b) //bn_curve C库函数
	if err != nil {
		log.Printf("NewIDFromBigInt %s\n", err)
		return nil
	}
	return id
}
