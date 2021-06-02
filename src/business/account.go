package business

import (
	"RocketTool/src/ecdsa/bls"
	"RocketTool/src/ecdsa/secp256k1"
	"RocketTool/src/ecdsa/sha3"
	"RocketTool/src/ecdsa/vrf"
	"RocketTool/src/model"
	"RocketTool/src/util"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
)

const (
	PubKeyLength = 65 //公钥字节长度，1 bytes curve, 64 bytes x,y。
	SecKeyLength = 97 //私钥字节长度，65 bytes pub, 32 bytes D。

	PREFIX = "0x"
)

type privateKey struct {
	key ecdsa.PrivateKey
}

type publicKey struct {
	key ecdsa.PublicKey
}

func CreateNewAccount() {
	r := rand.Reader
	var pk privateKey
	_pk, err := ecdsa.GenerateKey(secp256k1.S256(), r)
	if err == nil {
		pk.key = *_pk
	} else {
		panic(fmt.Sprintf("GenKey Failed, reason : %v.\n", err.Error()))
	}
	printAccountInfo(pk)
}

func newAccount() (string, string, privateKey) {
	r := rand.Reader
	var pk privateKey
	_pk, err := ecdsa.GenerateKey(secp256k1.S256(), r)
	if err == nil {
		pk.key = *_pk
	} else {
		panic(fmt.Sprintf("GenKey Failed, reason : %v.\n", err.Error()))
	}
	privateKeyStr := pk.getHexString()

	publicKey := pk.getPubKey()
	publicKeyStr := publicKey.getHexString()

	return privateKeyStr, publicKeyStr, pk
}

func printAccountInfo(privateKey privateKey) {

	publicKey := privateKey.getPubKey()
	address := publicKey.GetAddress()

	var miner model.Miner
	miner.Id = address.Bytes()

	secretSeed := util.RandFromBytes(address.Bytes())
	minerSecKey := *bls.NewSeckeyFromRand(secretSeed)
	minerPubKey := *bls.GeneratePubkey(minerSecKey)
	vrfPK, _, _ := vrf.VRFGenerateKey(bytes.NewReader(secretSeed.Bytes()))

	miner.PublicKey = minerPubKey.Serialize()
	miner.VrfPublicKey = vrfPK
	minerJson, _ := json.Marshal(miner)

	fmt.Println("Account info:")
	fmt.Println("PrivateKey:" + privateKey.getHexString())
	fmt.Println("MinerJson:" + string(minerJson))
}

//导入函数
func HexStringToSecKey(s string) (sk *privateKey) {
	if len(s) < len(PREFIX) || s[:len(PREFIX)] != PREFIX {
		return
	}
	buf, _ := hex.DecodeString(s[len(PREFIX):])
	sk = BytesToSecKey(buf)
	return
}

func BytesToSecKey(data []byte) (sk *privateKey) {
	//fmt.Printf("begin bytesToSecKey, len=%v, data=%v.\n", len(data), data)
	if len(data) < SecKeyLength {
		return nil
	}
	sk = new(privateKey)
	buf_pub := data[:PubKeyLength]
	buf_d := data[PubKeyLength:]
	sk.key.PublicKey = BytesToPublicKey(buf_pub).key
	sk.key.D = new(big.Int).SetBytes(buf_d)
	if sk.key.X != nil && sk.key.Y != nil && sk.key.D != nil {
		return sk
	}
	return nil
}

//从字节切片转换到公钥
func BytesToPublicKey(data []byte) (pk *publicKey) {
	pk = new(publicKey)
	pk.key.Curve = secp256k1.S256()
	//fmt.Printf("begin pub key unmarshal, len=%v, data=%v.\n", len(data), data)
	x, y := elliptic.Unmarshal(pk.key.Curve, data)
	if x == nil || y == nil {
		panic("unmarshal public key failed.")
	}
	pk.key.X = x
	pk.key.Y = y
	return
}

func (pk *privateKey) getHexString() string {
	buf := pk.toBytes()
	str := PREFIX + hex.EncodeToString(buf)
	return str
}

func (pk *privateKey) toBytes() []byte {
	buf := make([]byte, SecKeyLength)
	copy(buf[:PubKeyLength], pk.getPubKey().toBytes())
	d := pk.key.D.Bytes()
	if len(d) > 32 {
		panic("privateKey data length error: D length is more than 32!")
	}
	copy(buf[SecKeyLength-len(d):SecKeyLength], d)
	return buf
}

func (pk *privateKey) getPubKey() publicKey {
	var pubKey publicKey
	pubKey.key = pk.key.PublicKey
	return pubKey
}

func (pk publicKey) getHexString() string {
	buf := pk.toBytes()
	str := PREFIX + hex.EncodeToString(buf)
	return str
}

func (pk publicKey) toBytes() []byte {
	buf := elliptic.Marshal(pk.key.Curve, pk.key.X, pk.key.Y)
	return buf
}

func (pk publicKey) getID() []byte {
	x := pk.key.X.Bytes()
	y := pk.key.Y.Bytes()

	digest := make([]byte, 64)
	copy(digest[32-len(x):], x)
	copy(digest[64-len(y):], y)

	d := sha3.NewKeccak256()
	d.Write(digest)
	hash := d.Sum(nil)
	return hash
}

//由公钥萃取地址函数
func (pk publicKey) GetAddress() util.Address {
	addrBuf := pk.getID()
	return util.BytesToAddress(addrBuf[:])
}
