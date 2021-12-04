package business

import (
	"RocketTool/src/ecdsa/bls"
	"RocketTool/src/ecdsa/ed25519"
	"RocketTool/src/ecdsa/secp256k1"
	"RocketTool/src/model"
	"RocketTool/src/util"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
)

var (
	genesisGroupMemberNum uint64

	groupMemberList []*model.MinerInfo

	genesisGroupHeader *model.GroupHeader

	threshold int
)

func CreateGenesisGroup(groupMemberNum uint64) {
	genesisGroupMemberNum = groupMemberNum
	groupMemberList = make([]*model.MinerInfo, 0)
	createGroupMembers()

	groupMembers := make([]bls.ID, 0)
	for _, member := range groupMemberList {
		groupMembers = append(groupMembers, member.ID)
	}
	genesisGroupHeader = model.NewGenesisGroupHeader(groupMembers)

	mockGenSharePiece()

	groupPublicKey := mockGotAllSharePiece()
	showInfo(groupPublicKey, groupMembers)
}

func createGroupMembers() {
	var i uint64 = 0
	for ; i < genesisGroupMemberNum; i++ {
		miner := model.MinerInfo{}
		var sk privateKey
		var idStr string
		miner.PrivateKey, miner.PublicKey, idStr, sk = newAccount()
		miner.ID.SetHexString(idStr)

		miner.SecretSeed = util.RandFromBytes(sk.key.D.Bytes())
		miner.MinerSeckey = *bls.NewSeckeyFromRand(miner.SecretSeed)
		miner.MinerPublicKey = *bls.GeneratePubkey(miner.MinerSeckey)

		miner.VrfPK, miner.VrfSK, _ = ed25519.GenerateKey(&miner)
		miner.ReceivedSharePiece = make([]*model.SharePiece, 0)

		groupMemberList = append(groupMemberList, &miner)
	}
}

func newAccount() (string, string, string, privateKey) {
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

	idStr := PREFIX + hex.EncodeToString(publicKey.getID())
	return privateKeyStr, publicKeyStr, idStr, pk
}

func mockGenSharePiece() {
	threshold = getThreshold()
	for i := 0; i < len(groupMemberList); i++ {
		miner := groupMemberList[i]
		genSharePiece(threshold, *miner)
	}
}

func genSharePiece(threshold int, minerInfo model.MinerInfo) {
	secretList := make([]bls.Seckey, threshold)
	for i := 0; i < threshold; i++ {
		secretList[i] = *bls.NewSeckeyFromRand(minerInfo.SecretSeed.Deri(i))
	}

	seedPubkey := getSeedPubKey(minerInfo)
	for i := 0; i < len(groupMemberList); i++ {
		miner := groupMemberList[i]

		sharePiece := new(model.SharePiece)
		sharePiece.Pub = seedPubkey
		sharePiece.Share = *bls.ShareSeckey(secretList, miner.ID)

		miner.ReceivedSharePiece = append(miner.ReceivedSharePiece, sharePiece)
	}
}

func mockGotAllSharePiece() bls.Pubkey {
	signPublicKeyList := make([]bls.Pubkey, 0)
	for index, member := range groupMemberList {
		receivedShareList := make([]bls.Seckey, 0)
		for _, sharePiece := range member.ReceivedSharePiece {
			receivedShareList = append(receivedShareList, sharePiece.Share)
			if index == 0 {
				signPublicKeyList = append(signPublicKeyList, sharePiece.Pub)
			}
		}
		signPrivateKeyInGroup := bls.AggregateSeckeys(receivedShareList)
		groupMemberList[index].SignPrivateKeyInGroup = signPrivateKeyInGroup
	}
	groupPublicKey := bls.AggregatePubkeys(signPublicKeyList)
	fmt.Printf("Group pubkey:%s\n", groupPublicKey.GetHexString())

	memberSignMap := make(map[string]bls.Signature, 0)
	for _, member := range groupMemberList {
		sign := bls.Sign(*member.SignPrivateKeyInGroup, genesisGroupHeader.Hash[:])
		memberSignMap[member.ID.GetHexString()] = sign
	}
	groupSign := bls.RecoverGroupSignature(memberSignMap, threshold)

	verifyResult := bls.VerifySig(*groupPublicKey, genesisGroupHeader.Hash[:], *groupSign)
	if !verifyResult {
		panic("Group sign verify failed! Please contact the developer.")
	}
	return *groupPublicKey
}

func showInfo(groupPublicKey bls.Pubkey, groupMembers []bls.ID) {
	showGroupMemberInfo()
	groupID := bls.NewIDFromPubkey(groupPublicKey)
	showGenesisGroupInfo(groupMembers, *groupID, groupPublicKey)
	showJoinedGroupInfo(*groupID, groupPublicKey)
}

func showGroupMemberInfo() {
	fmt.Println("Genesis member info:")
	for _, member := range groupMemberList {
		fmt.Printf("PrivateKey:%s\n", member.PrivateKey)
		fmt.Printf("SignSecKey:%s\n", member.SignPrivateKeyInGroup.GetHexString())
		fmt.Printf("ID:%s\n\n", member.ID.GetHexString())
	}
}

func showGenesisGroupInfo(groupMembers []bls.ID, groupID bls.ID, groupPublicKey bls.Pubkey) {
	groupInitInfo := model.GroupInitInfo{GroupHeader: genesisGroupHeader, GroupMembers: groupMembers}

	memberIndexMap := make(map[string]int, 0)
	vrfPublicKeyList := make([]ed25519.PublicKey, 0)
	publicKeyList := make([]bls.Pubkey, 0)

	for index, member := range groupMemberList {
		memberIndexMap[member.ID.GetHexString()] = index

		vrfPublicKeyList = append(vrfPublicKeyList, member.VrfPK)
		publicKeyList = append(publicKeyList, member.MinerPublicKey)
	}
	groupInfo := model.GroupInfo{GroupID: groupID, GroupPK: groupPublicKey, GroupInitInfo: &groupInitInfo, MemberIndexMap: memberIndexMap}

	genesisGroup := model.GenesisGroup{GroupInfo: groupInfo, VrfPubkey: vrfPublicKeyList, Pubkeys: publicKeyList}
	groupBytes, _ := json.Marshal(genesisGroup)
	fmt.Println("Gnenesis group info:\n" + string(groupBytes) + "\n")
}

func showJoinedGroupInfo(groupID bls.ID, groupPublicKey bls.Pubkey) {
	joinedGroupInfo := model.JoinedGroupInfo{GroupHash: genesisGroupHeader.Hash, GroupID: groupID, GroupPK: groupPublicKey}

	memberSignPubkeyMap := make(map[string]bls.Pubkey, 0)
	for _, member := range groupMemberList {
		signPublicKeyInGroup := bls.GeneratePubkey(*member.SignPrivateKeyInGroup)
		memberSignPubkeyMap[member.ID.GetHexString()] = *signPublicKeyInGroup
	}
	joinedGroupInfo.MemberSignPubkeyMap = memberSignPubkeyMap
	joinedGroupByte, _ := json.Marshal(joinedGroupInfo)
	fmt.Println("Joined group info:\n" + string(joinedGroupByte))
}

func getSeedPubKey(minerInfo model.MinerInfo) bls.Pubkey {
	seedSecKey := bls.NewSeckeyFromRand(minerInfo.SecretSeed.Deri(0))
	return *bls.GeneratePubkey(*seedSecKey)
}

func getThreshold() int {
	return int(math.Ceil(float64(genesisGroupMemberNum*51) / 100))
}
