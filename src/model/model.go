package model

import (
	"RocketTool/src/ecdsa/bls"
	"RocketTool/src/ecdsa/ed25519"
	"RocketTool/src/util"
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

type Miner struct {
	Id           []byte `json:"id,omitempty"`
	PublicKey    []byte `json:"publicKey,omitempty"`
	VrfPublicKey []byte `json:"vrfPublicKey,omitempty"`
}
