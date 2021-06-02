package business

import (
	"RocketTool/src/ecdsa/bls"
	"RocketTool/src/ecdsa/ed25519"
	"RocketTool/src/model"
	"RocketTool/src/util"
	"fmt"
	"testing"
)

func TestCreateNewAccount(t *testing.T) {
	pkString := "0x0411c6362e7ece1fcbbe98085bcc8410749c2bebf546da7d7406e1e9a20afdabe29eb6826e1ea25c3d0e8c2db953661d5e37c1f5f71dc8fbd5d0b7422e3f0aeff320f7f665561eae7693ade2f9592530c2f9b67b1d6cc2897b0c71b7d7b8d02a3a"
	pk := HexStringToSecKey(pkString)

	fmt.Println(pk)

	miner := model.MinerInfo{}
	miner.SecretSeed = util.RandFromBytes(pk.key.D.Bytes())
	miner.MinerSeckey = *bls.NewSeckeyFromRand(miner.SecretSeed)
	miner.MinerPublicKey = *bls.GeneratePubkey(miner.MinerSeckey)

	idBytes := pk.getPubKey().getID()
	miner.ID.Deserialize(idBytes)

	miner.VrfPK, miner.VrfSK, _ = ed25519.GenerateKey(&miner)

	fmt.Println(miner)

}
