package business

import (
	"RocketTool/src/ecdsa/bls"
	"RocketTool/src/ecdsa/vrf"
	"RocketTool/src/util"
)

// 矿工信息
type MinerInfo struct {
	// 矿工签名公钥，用于建组、出块等消息的签名及验证
	PubKey bls.Pubkey

	// 矿工ID
	ID bls.ID

	// VRF公钥，用于验证VRFProve
	VrfPK vrf.VRFPublicKey

	Stake     uint64
	MinerType byte

	ApplyHeight uint64
	AbortHeight uint64
}

type SelfMinerInfo struct {
	SecretSeed util.Rand //私密随机数
	SecKey     bls.Seckey
	VrfSK      vrf.VRFPrivateKey

	MinerInfo
}

func NewSelfMinerInfo(privateKey privateKey) SelfMinerInfo {
	var mi SelfMinerInfo
	mi.SecretSeed = util.RandFromBytes(privateKey.key.D.Bytes())
	mi.SecKey = *bls.NewSeckeyFromRand(mi.SecretSeed)
	mi.PubKey = *bls.GeneratePubkey(mi.SecKey)
	idBytes := privateKey.getPubKey().getID()
	mi.ID = bls.ID{}
	mi.ID.Deserialize(idBytes)

	var err error
	mi.VrfPK, mi.VrfSK, err = vrf.VRFGenerateKey(&mi)
	if err != nil {
		panic("generate vrf key error, err=" + err.Error())
	}
	return mi
}

func (md *SelfMinerInfo) Read(p []byte) (n int, err error) {
	bs := md.SecretSeed.Bytes()
	if p == nil || len(p) < len(bs) {
		p = make([]byte, len(bs))
	}
	copy(p, bs)
	return len(bs), nil
}
