package business

import (
	"RocketTool/src/ecdsa/secp256k1"
	"RocketTool/src/ecdsa/sha3"
	"RocketTool/src/model"
	"RocketTool/src/util"
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

	CHAIN_ID = "2025" // Rangers mainnet
	PREFIX = "0x"
)

type privateKey struct {
	key ecdsa.PrivateKey
}

type publicKey struct {
	key ecdsa.PublicKey
}

func CreateNewAccount(nodeType int, privateKeyString string) {
	var pk privateKey
	if 0 == len(privateKeyString) {
		_pk, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
		if err == nil {
			pk.key = *_pk
		} else {
			panic(fmt.Sprintf("GenKey Failed, reason : %v.\n", err.Error()))
		}
	} else {
		pk = *HexStringToSecKey(privateKeyString)

	}

	printAccountInfo(pk, nodeType)
}

func printAccountInfo(privateKey privateKey, nodeType int) {
	selfMinerInfo := NewSelfMinerInfo(privateKey)
	var miner model.Miner
	miner.Id = selfMinerInfo.ID.Serialize()
	miner.PublicKey = selfMinerInfo.PubKey.Serialize()
	miner.VrfPublicKey = selfMinerInfo.VrfPK
	minerJson, _ := json.Marshal(miner)

	fmt.Println("Account info:")
	fmt.Println("PrivateKey:" + privateKey.getHexString())
	fmt.Println("MinerJson:" + string(minerJson))
	if -1 != nodeType {
		publicKey := privateKey.getPubKey()
		address := publicKey.GetAddress()
		printMinerApplyTx(nodeType, address.GetHexString(), privateKey, miner)
	}
}

func printMinerApplyTx(nodeType int, source string, privateKey privateKey, miner model.Miner) {
	{
		tx := model.Transaction{Type: 2, Source: source, Target: source, Time: time.Now().String(), ChainId: CHAIN_ID}

		if 1 == nodeType {
			miner.Stake = 2000
		} else {
			miner.Stake = 400
		}
		miner.Type = nodeType
		applyData, _ := json.Marshal(miner)
		fmt.Printf("data:%v\n",string(applyData))

		tx.Data = string(applyData)
		tx.Hash = tx.GenHash()

		sign := privateKey.Sign(tx.Hash.Bytes())
		tx.Sign = &sign

		fmt.Printf("tx: %s\n\n", tx.ToTxJson().ToString())
	}
}

//导入函数
//func HexStringToSecKey(s string) (sk *privateKey) {
//	if len(s) < len(PREFIX) || s[:len(PREFIX)] != PREFIX {
//		return
//	}
//	buf, _ := hex.DecodeString(s[len(PREFIX):])
//	sk = BytesToSecKey(buf)
//	return
//}

//导入函数
func HexStringToSecKey(s string) (sk *privateKey) {
	if len(s) < len(PREFIX) || s[:len(PREFIX)] != PREFIX {
		return
	}
	sk = new(privateKey)
	sk.key.D = new(big.Int).SetBytes(util.FromHex(s))
	sk.key.PublicKey.Curve = getDefaultCurve()
	sk.key.PublicKey.X, sk.key.PublicKey.Y = getDefaultCurve().ScalarBaseMult(sk.key.D.Bytes())
	return
}

func getDefaultCurve() elliptic.Curve {
	return secp256k1.S256()
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
	buf := pk.key.D.Bytes()
	str := PREFIX + hex.EncodeToString(buf)
	return str
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
