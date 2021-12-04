package model

import (
	"RocketTool/src/ecdsa/bls"
	"RocketTool/src/ecdsa/ed25519"
	"RocketTool/src/util"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
)

type MinerInfo struct {
	PrivateKey string
	PublicKey  string
	ID         bls.ID

	SecretSeed     util.Rand
	MinerSeckey    bls.Seckey
	MinerPublicKey bls.Pubkey

	VrfSK ed25519.PrivateKey
	VrfPK ed25519.PublicKey

	ReceivedSharePiece    []*SharePiece
	SignPrivateKeyInGroup *bls.Seckey
}

func (miner *MinerInfo) Read(p []byte) (n int, err error) {
	bs := miner.SecretSeed.Bytes()
	if p == nil || len(p) < len(bs) {
		p = make([]byte, len(bs))
	}
	copy(p, bs)
	return len(bs), nil
}

type SharePiece struct {
	Share bls.Seckey
	Pub   bls.Pubkey
}

type HexBytes []byte

func (h *HexBytes) UnmarshalJSON(b []byte) error {
	if 2 > len(b) {
		return fmt.Errorf("length error, %d", len(b))
	}
	res := string(b[1 : len(b)-1])
	*h = util.FromHex(res)
	return nil
}

func (h HexBytes) MarshalJSON() ([]byte, error) {
	res := fmt.Sprintf("\"%s\"", util.ToHex(h))
	return []byte(res), nil
}

type Miner struct {
	Id           HexBytes `json:"id,omitempty"`
	PublicKey    HexBytes `json:"publicKey,omitempty"`
	VrfPublicKey []byte   `json:"vrfPublicKey,omitempty"`

	Type int `json:"type,omitempty"`

	// 质押数
	Stake uint64 `json:"stake,omitempty"`

	ApplyHeight uint64
	AbortHeight uint64 `json:"-"`

	// 当前状态
	Status byte
}

type Sign struct {
	R     big.Int
	S     big.Int
	Recid byte
}

//Sign必须65 bytes
func (s Sign) Bytes() []byte {
	rb := s.R.Bytes()
	sb := s.S.Bytes()
	r := make([]byte, 65)
	copy(r[32-len(rb):32], rb)
	copy(r[64-len(sb):64], sb)
	r[64] = s.Recid
	return r
}

func (s Sign) GetHexString() string {
	buf := s.Bytes()
	str := "0x" + hex.EncodeToString(buf)
	return str
}

type Transaction struct {
	Source string // 用户id
	Target string // 游戏id
	Type   int32  // 场景id
	Time   string

	Data          string // 状态机入参
	ExtraData     string // 在rocketProtocol里，用于转账。包括余额转账、FT转账、NFT转账
	ExtraDataType int32

	Hash util.Hash
	Sign *Sign

	Nonce           uint64 // 用户级别nonce
	RequestId       uint64 // 消息编号 由网关添加
	SocketRequestId string // websocket id，用于客户端标示请求id，方便回调处理
}

func (tx *Transaction) GenHash() util.Hash {
	if nil == tx {
		return util.Hash{}
	}
	buffer := bytes.Buffer{}

	buffer.Write([]byte(tx.Data))
	buffer.Write([]byte(strconv.FormatUint(tx.Nonce, 10)))
	buffer.Write([]byte(tx.Source))
	buffer.Write([]byte(tx.Target))
	buffer.Write([]byte(strconv.Itoa(int(tx.Type))))
	buffer.Write([]byte(tx.Time))
	buffer.Write([]byte(tx.ExtraData))
	return util.BytesToHash(util.Sha256(buffer.Bytes()))
}

func (tx Transaction) ToTxJson() TxJson {
	txJson := TxJson{Source: tx.Source, Target: tx.Target, Type: tx.Type, Time: tx.Time,
		Data: tx.Data, ExtraData: tx.ExtraData, Nonce: tx.Nonce,
		Hash: tx.Hash.String(), RequestId: tx.RequestId, SocketRequestId: tx.SocketRequestId}

	if tx.Sign != nil {
		txJson.Sign = tx.Sign.GetHexString()
	}
	return txJson
}

type TxJson struct {
	// 用户id
	Source string `json:"source"`
	// 游戏id
	Target string `json:"target"`
	// 场景id
	Type int32  `json:"type"`
	Time string `json:"time,omitempty"`

	// 入参
	Data      string `json:"data,omitempty"`
	ExtraData string `json:"extraData,omitempty"`

	Hash string `json:"hash,omitempty"`
	Sign string `json:"sign,omitempty"`

	Nonce           uint64 `json:"nonce,omitempty"`
	RequestId       uint64
	SocketRequestId string `json:"socketRequestId,omitempty"`
}

func (txJson TxJson) ToString() string {
	byte, err := json.Marshal(txJson)
	if err != nil {
		return ""
	}
	return string(byte)
}
