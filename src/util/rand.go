package util

import (
	"crypto/rand"
	"math/big"
	"strconv"
)

const RandLength = 32

type Rand [RandLength]byte

func NewRand() (r Rand) {
	b := make([]byte, RandLength)
	rand.Read(b)
	return RandFromBytes(b)
}

func RandFromBytes(b ...[]byte) (r Rand) {
	HashBytes(b...).Sum(r[:0])
	return
}

func (r Rand) Deri(vi ...int) Rand {
	return r.ders(mapItoa(vi)...)
}

func (r Rand) Bytes() []byte {
	return r[:]
}

func (r Rand) derivedRand(x ...[]byte) Rand {
	ri := r
	for _, xi := range x { //遍历多维字节数组
		HashBytes(ri.Bytes(), xi).Sum(ri[:0]) //哈希叠加计算
	}
	return ri
}

func (r Rand) ders(s ...string) Rand {
	return r.derivedRand(mapStringToBytes(s)...)
}

func mapStringToBytes(x []string) [][]byte {
	y := make([][]byte, len(x))
	for k, xi := range x {
		y[k] = []byte(xi)
	}
	return y
}

func mapItoa(x []int) []string {
	y := make([]string, len(x))
	for k, xi := range x {
		y[k] = strconv.Itoa(xi)
	}
	return y
}

func (r Rand) RandomPerm(n int, k int) []int {
	l := make([]int, n)
	for i := range l {
		l[i] = i
	}
	for i := 0; i < k; i++ {
		j := r.Deri(i).modulo(n-i) + i
		l[i], l[j] = l[j], l[i]
	}
	return l[:k]
}

func (r Rand) modulo(n int) int {
	b := big.NewInt(0)
	b.SetBytes(r.Bytes())          //随机数转换成big.Int
	b.Mod(b, big.NewInt(int64(n))) //对n求模
	return int(b.Int64())
}
