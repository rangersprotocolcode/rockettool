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
	"time"
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

func CreateNewAccount(nodeType int) {
	r := rand.Reader
	var pk privateKey
	_pk, err := ecdsa.GenerateKey(secp256k1.S256(), r)
	if err == nil {
		pk.key = *_pk
	} else {
		panic(fmt.Sprintf("GenKey Failed, reason : %v.\n", err.Error()))
	}
	printAccountInfo(pk, nodeType)
}

func printAccountInfo(privateKey privateKey, nodeType int) {

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
	if -1 != nodeType {
		printMinerApplyTx(nodeType, "0x"+hex.EncodeToString(address[:]), string(minerJson))
	}
}

func printMinerApplyTx(nodeType int, target, data string) {
	{
		source := "0x6420e467c77514e09471a7d84e0552c13b5e97192f523c05d3970d7ee23bf443"
		tx := model.Transaction{Type: 2, Source: source, Target: target, Time: time.Now().String()}

		//data := `{"id":"mlrcS4PtQnL4rwxGaGqThwE5GuNXa3eJHiq050OPRC4=","publicKey":"BOu0RbvBDBlVUySzb+ojoE7BTO67yhYQWdOvqClYG+Qu11SFY79i1lDou9VkPfnpX0KPhlvtpTIIK3IIR2K1meM=","vrfPublicKey":"Dw7zNJeE4wj+diK2c/P+9raL6R72SY1ySbleYVihJtU="}`
		var obj = model.Miner{}
		err := json.Unmarshal([]byte(data), &obj)
		if err != nil {
			fmt.Printf("ummarshal error:%v", err)
		}

		if 1 == nodeType {
			obj.Stake = 1250
		} else {
			obj.Stake = 250
		}
		obj.Type = nodeType

		applyData, _ := json.Marshal(obj)
		//fmt.Printf("data:%v\n",string(applyData))

		tx.Data = string(applyData)
		tx.Hash = tx.GenHash()

		privateKeyStr := "0x040a0c4baa2e0b927a2b1f6f93b317c320d4aa3a5b54c0a83f5872c23155dcf1455fb015a7699d4ef8491cc4c7a770e580ab1362a0e3af9f784dd2485cfc9ba7c1e7260a418579c2e6ca36db4fe0bf70f84d687bdf7ec6c0c181b43ee096a84aea"
		privateKey := HexStringToSecKey(privateKeyStr)
		sign := privateKey.Sign(tx.Hash.Bytes())
		tx.Sign = &sign

		fmt.Printf("tx: %s\n\n", tx.ToTxJson().ToString())
	}
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

//私钥签名函数
func (pk *privateKey) Sign(hash []byte) model.Sign {
	var sign model.Sign
	sig, err := secp256k1.Sign(hash, pk.key.D.Bytes())
	if err == nil {
		if len(sig) != 65 {
			fmt.Printf("secp256k1 sign wrong! hash = %v\n", hash)
		}
		sign = *BytesToSign(sig)
	} else {
		panic(fmt.Sprintf("Sign Failed, reason : %v.\n", err.Error()))
	}

	return sign
}

//Sign必须65 bytes
func BytesToSign(b []byte) *model.Sign {
	if len(b) == 65 {
		var r, s big.Int
		br := b[:32]
		r = *r.SetBytes(br)

		sr := b[32:64]
		s = *s.SetBytes(sr)

		recid := b[64]
		return &model.Sign{r, s, recid}
	} else {
		//这里组签名暂不处理
		return nil
	}
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
