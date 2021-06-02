package util

import (
	"crypto/sha256"
	"golang.org/x/crypto/sha3"
	"hash"
	"math/rand"
	"reflect"
	"sync"
)

const MaxUint64 = 1<<64 - 1

const AddressLength = 20
const HashLength = 32

var (
	hashT = reflect.TypeOf(Hash{})
)

type Hash [HashLength]byte
type Address [AddressLength]byte

//构造函数族
func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

//赋值函数，如b超出a的容量则截取后半部分
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[:], b[:])
}

func (a Address) Bytes() []byte { return a[:] }

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

func (h Hash) Bytes() []byte { return h[:] }

// UnmarshalText parses a hash in hex syntax.
func (h *Hash) UnmarshalText(input []byte) error {
	return UnmarshalFixedText("Hash", input, h[:])
}

// UnmarshalJSON parses a hash in hex syntax.
func (h *Hash) UnmarshalJSON(input []byte) error {
	return UnmarshalFixedJSON(hashT, input, h[:])
}

// MarshalText returns the hex representation of h.
func (h Hash) MarshalText() ([]byte, error) {
	return bytes(h[:]).MarshalText()
}

// Sets the hash to the value of b. If b is larger than len(h), 'b' will be cropped (from the left).
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:] //截取右边部分
	}

	copy(h[HashLength-len(b):], b)
}

// Set string `s` to h. If s is larger than len(h) s will be cropped (from left) to fit.
func (h *Hash) SetString(s string) { h.SetBytes([]byte(s)) }

// Sets h from other
func (h *Hash) Set(other Hash) {
	for i, v := range other {
		h[i] = v
	}
}

// Generate implements testing/quick.Generator.
func (h Hash) Generate(rand *rand.Rand, size int) reflect.Value {
	m := rand.Intn(len(h))            //m为0-len(h)之间的伪随机数
	for i := len(h) - 1; i > m; i-- { //从高位到m之间进行遍历
		h[i] = byte(rand.Uint32()) //rand.Uint32为32位非负伪随机数
	}
	return reflect.ValueOf(h)
}

var hasherPool = sync.Pool{
	New: func() interface{} {
		return sha256.New()
	},
}

// 计算sha256
func Sha256(blockByte []byte) []byte {
	hasher := hasherPool.Get().(hash.Hash)
	hasher.Reset()
	defer hasherPool.Put(hasher)

	hasher.Write(blockByte)
	return hasher.Sum(nil)
}

func HashBytes(b ...[]byte) hash.Hash {
	d := sha3.New256()
	for _, bi := range b {
		d.Write(bi)
	}
	return d
}
