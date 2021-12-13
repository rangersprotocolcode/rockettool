package business

import (
	"RocketTool/src/model"
	"RocketTool/src/util"
	"encoding/json"
	"fmt"
	"time"
)

func GenerateMinerRewardAccountTx(privateKeyStr, id, account string) {
	fmt.Printf("generate tx for pk: %s, id: %s, account: %s, please check them\n", privateKeyStr, id, account)

	privateKey := HexStringToSecKey(privateKeyStr)
	source := privateKey.getPubKey().GetAddress().String()
	tx := model.Transaction{Type: 6, Source: source, Time: time.Now().String(), ChainId: "8888"}
	var miner model.Miner
	miner.Id = util.FromHex(id)
	miner.Account = util.FromHex(account)
	data, _ := json.Marshal(miner)
	tx.Data = string(data)
	tx.Hash = tx.GenHash()

	sign := privateKey.Sign(tx.Hash.Bytes())
	tx.Sign = &sign

	fmt.Printf("\n\nraw transaction: %s\n", tx.ToTxJson().ToString())
}
