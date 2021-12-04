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
	pkString := "0x041aa9c9a10319ed4bbc50833dcd5084213978db8b81167c39054de25e6ec2aa66835a629ad94aed80e878754724809f07be09247126f3cf4eadb225c7a2f6c764d7f5d173593eff81a50f7d8ea345bbc543ad8e356e75975e87114438c8f4eaf4"
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

func TestCreateNewAddress(t *testing.T) {
	pkString := "0x041aa9c9a10319ed4bbc50833dcd5084213978db8b81167c39054de25e6ec2aa66835a629ad94aed80e878754724809f07be09247126f3cf4eadb225c7a2f6c764d7f5d173593eff81a50f7d8ea345bbc543ad8e356e75975e87114438c8f4eaf4"
	pk := HexStringToSecKey(pkString)

	addr := pk.getPubKey().GetAddress()

	fmt.Println(util.ToHex(addr.Bytes()))

}
