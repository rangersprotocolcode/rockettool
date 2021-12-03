package model

import (
	"RocketTool/src/ecdsa/bls"
	"RocketTool/src/ecdsa/ed25519"
	"RocketTool/src/util"
	"bytes"
	"encoding/binary"
	"golang.org/x/crypto/sha3"
	"time"
)

type GenesisGroup struct {
	GroupInfo GroupInfo
	VrfPubkey []ed25519.PublicKey
	Pubkeys   []bls.Pubkey
}

type GroupInfo struct {
	GroupID       bls.ID
	GroupPK       bls.Pubkey
	GroupInitInfo *GroupInitInfo

	MemberIndexMap map[string]int
	ParentGroupID  bls.ID
	PrevGroupID    bls.ID
}

type GroupInitInfo struct {
	GroupHeader     *GroupHeader
	ParentGroupSign bls.Signature
	GroupMembers    []bls.ID
}

type JoinedGroupInfo struct {
	GroupHash util.Hash
	GroupID   bls.ID
	GroupPK   bls.Pubkey

	SignSecKey          bls.Seckey
	MemberSignPubkeyMap map[string]bls.Pubkey
}

type GroupHeader struct {
	Hash          util.Hash
	Parent        []byte
	PreGroup      []byte
	Authority     uint64
	Name          string
	BeginTime     time.Time
	MemberRoot    util.Hash
	CreateHeight  uint64
	ReadyHeight   uint64
	WorkHeight    uint64
	DismissHeight uint64
	Extends       string
}

func NewGenesisGroupHeader(memIds []bls.ID) *GroupHeader {
	gh := &GroupHeader{
		Name:          "Rangers Protocol Genesis Group",
		Authority:     777,
		BeginTime:     time.Now(),
		CreateHeight:  0,
		ReadyHeight:   1,
		WorkHeight:    0,
		DismissHeight: util.MaxUint64,
		MemberRoot:    genGroupMemberRoot(memIds),
		Extends:       "",
	}

	gh.Hash = gh.genHash()
	return gh
}

func (gh *GroupHeader) genHash() util.Hash {
	buf := bytes.Buffer{}
	buf.Write(gh.Parent)
	buf.Write(gh.PreGroup)
	buf.Write(uint64ToByte(gh.Authority))
	buf.WriteString(gh.Name)

	buf.Write(gh.MemberRoot.Bytes())
	buf.Write(uint64ToByte(gh.CreateHeight))
	buf.Write(uint64ToByte(gh.ReadyHeight))
	buf.Write(uint64ToByte(gh.WorkHeight))
	buf.Write(uint64ToByte(gh.DismissHeight))
	buf.WriteString(gh.Extends)
	return util.BytesToHash(util.Sha256(buf.Bytes()))
}

func genGroupMemberRoot(ids []bls.ID) util.Hash {
	data := bytes.Buffer{}
	for _, m := range ids {
		data.Write(m.Serialize())
	}
	return data2CommonHash(data.Bytes())
}

func data2CommonHash(data []byte) util.Hash {
	var h util.Hash
	sha3_hash := sha3.Sum256(data)
	if len(sha3_hash) == util.HashLength {
		copy(h[:], sha3_hash[:])
	} else {
		panic("Data2Hash failed, size error.")
	}
	return h
}

func uint64ToByte(i uint64) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, i)
	return buf.Bytes()
}
