package LICY_BLC

import (
	"encoding/hex"
	"log"
	"github.com/boltdb/bolt"
	"bytes"
)

const utxoTableName  = "utxoTableName"

type Licy_UTXOSet struct {
	Licy_Blockchain *Licy_Blockchain
}

// 重置数据库表
func (utxoSet *Licy_UTXOSet) Licy_ResetUTXOSet()  {

	err := utxoSet.Licy_Blockchain.Licy_DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(utxoTableName))

		if b != nil {


			err := tx.DeleteBucket([]byte(utxoTableName))

			if err!= nil {
				log.Panic(err)
			}

		}

		b ,_ = tx.CreateBucket([]byte(utxoTableName))
		if b != nil {

			//[string]*TXOutputs
			txOutputsMap := utxoSet.Licy_Blockchain.Licy_FindUTXOMap()


			for keyHash,outs := range txOutputsMap {

				txHash,_ := hex.DecodeString(keyHash)

				b.Put(txHash,outs.Licy_Serialize())

			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

}

func (utxoSet *Licy_UTXOSet) Licy_findUTXOForAddress(address string) []*Licy_UTXO{


	var utxos []*Licy_UTXO

	utxoSet.Licy_Blockchain.Licy_DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(utxoTableName))

		// 游标
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			txOutputs := Licy_DeserializeTXOutputs(v)

			for _,utxo := range txOutputs.Licy_UTXOS  {

				if utxo.Licy_Output.Licy_UnLockPubKeyWithAddress(address) {
					utxos = append(utxos,utxo)
				}
			}
		}

		return nil
	})

	return utxos
}




func (utxoSet *Licy_UTXOSet) Licy_GetBalance(address string) int64 {

	UTXOS := utxoSet.Licy_findUTXOForAddress(address)

	var amount int64

	for _,utxo := range UTXOS  {
		amount += utxo.Licy_Output.Licy_Value
	}

	return amount
}


// 返回要凑多少钱，对应TXOutput的TX的Hash和index
func (utxoSet *Licy_UTXOSet) Licy_FindUnPackageSpendableUTXOS(from string, txs []*Licy_Transaction) []*Licy_UTXO {

	var unUTXOs []*Licy_UTXO

	spentTXOutputs := make(map[string][]int)

	//{hash:[0]}

	for _,tx := range txs {

		if tx.Licy_IsCoinbaseTransaction() == false {
			for _, in := range tx.Licy_Vins {
				//是否能够解锁
				publicKeyHash := Licy_Base58Decode([]byte(from))

				ripemd160Hash := publicKeyHash[1:len(publicKeyHash) - 4]
				if in.Licy_UnLockRipemd160Hash(ripemd160Hash) {

					key := hex.EncodeToString(in.Licy_TxHash)

					spentTXOutputs[key] = append(spentTXOutputs[key], in.Licy_Index)
				}

			}
		}
	}


	for _,tx := range txs {

	Work1:
		for index,out := range tx.Licy_Vouts {

			if out.Licy_UnLockPubKeyWithAddress(from) {

				if len(spentTXOutputs) == 0 {
					utxo := &Licy_UTXO{tx.Licy_TxHash, index, out}
					unUTXOs = append(unUTXOs, utxo)
				} else {
					for hash,indexArray := range spentTXOutputs {

						txHashStr := hex.EncodeToString(tx.Licy_TxHash)

						if hash == txHashStr {

							var isUnSpentUTXO bool

							for _,outIndex := range indexArray {

								if index == outIndex {
									isUnSpentUTXO = true
									continue Work1
								}

								if isUnSpentUTXO == false {
									utxo := &Licy_UTXO{tx.Licy_TxHash, index, out}
									unUTXOs = append(unUTXOs, utxo)
								}
							}
						} else {
							utxo := &Licy_UTXO{tx.Licy_TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}
		}
	}
	return unUTXOs
}

func (utxoSet *Licy_UTXOSet) Licy_FindSpendableUTXOS(from string,amount int64,txs []*Licy_Transaction) (int64,map[string][]int)  {

	unPackageUTXOS := utxoSet.Licy_FindUnPackageSpendableUTXOS(from,txs)

	spentableUTXO := make(map[string][]int)

	var money int64 = 0

	for _,Licy_UTXO := range unPackageUTXOS {

		money += Licy_UTXO.Licy_Output.Licy_Value;
		txHash := hex.EncodeToString(Licy_UTXO.Licy_TxHash)
		spentableUTXO[txHash] = append(spentableUTXO[txHash],Licy_UTXO.Licy_Index)
		if money >= amount{
			return  money,spentableUTXO
		}
	}
	// 钱还不够
	utxoSet.Licy_Blockchain.Licy_DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoTableName))
		if b != nil {
			c := b.Cursor()
		UTXOBREAK:
			for k, v := c.First(); k != nil; k, v = c.Next() {
				txOutputs := Licy_DeserializeTXOutputs(v)
				for _,utxo := range txOutputs.Licy_UTXOS {
					money += utxo.Licy_Output.Licy_Value
					txHash := hex.EncodeToString(utxo.Licy_TxHash)
					spentableUTXO[txHash] = append(spentableUTXO[txHash],utxo.Licy_Index)
					if money >= amount {
						break UTXOBREAK;
					}
				}
			}
		}
		return nil
	})
	if money < amount{
		log.Panic("余额不足......")
	}
	return  money,spentableUTXO
}

// 更新
func (utxoSet *Licy_UTXOSet) Licy_Update()  {

	// 最新的Block
	block := utxoSet.Licy_Blockchain.Iterator().Next()
	// utxoTable
	ins := []*Licy_TxInput{}
	outsMap := make(map[string]*Licy_TxOutputs)
	// 找到所有我要删除的数据
	for _,tx := range block.Licy_Txs {
		for _,in := range tx.Licy_Vins {
			ins = append(ins,in)
		}
	}
	for _,tx := range block.Licy_Txs  {
		utxos := []*Licy_UTXO{}
		for index,out := range tx.Licy_Vouts  {
			isSpent := false
			for _,in := range ins  {
				if in.Licy_Index == index && bytes.Compare(tx.Licy_TxHash ,in.Licy_TxHash) == 0 && bytes.Compare(out.Licy_Ripemd160Hash,Ripemd160Hash(in.Licy_PublicKey)) == 0 {
					isSpent = true
					continue
				}
			}
			if isSpent == false {
				utxo := &Licy_UTXO{tx.Licy_TxHash,index,out}
				utxos = append(utxos,utxo)
			}
		}
		if len(utxos) > 0 {
			txHash := hex.EncodeToString(tx.Licy_TxHash)
			outsMap[txHash] = &Licy_TxOutputs{utxos}
		}
	}
	err := utxoSet.Licy_Blockchain.Licy_DB.Update(func(tx *bolt.Tx) error{
		b := tx.Bucket([]byte(utxoTableName))
		if b != nil {
			// 删除
			for _,in := range ins {
				txOutputsBytes := b.Get(in.Licy_TxHash)
				if len(txOutputsBytes) == 0 {
					continue
				}
				//fmt.Println("DeserializeTXOutputs")
				//fmt.Println(txOutputsBytes)
				txOutputs := Licy_DeserializeTXOutputs(txOutputsBytes)
				//fmt.Println(txOutputs)
				UTXOS := []*Licy_UTXO{}
				// 判断是否需要
				isNeedDelete := false
				for _,utxo := range txOutputs.Licy_UTXOS  {
					if in.Licy_Index == utxo.Licy_Index && bytes.Compare(utxo.Licy_Output.Licy_Ripemd160Hash,Ripemd160Hash(in.Licy_PublicKey)) == 0 {
						isNeedDelete = true
					} else {
						UTXOS = append(UTXOS,utxo)
					}
				}
				if isNeedDelete {
					b.Delete(in.Licy_TxHash)
					if len(UTXOS) > 0 {
						preTXOutputs := outsMap[hex.EncodeToString(in.Licy_TxHash)]
						preTXOutputs.Licy_UTXOS = append(preTXOutputs.Licy_UTXOS,UTXOS...)
						outsMap[hex.EncodeToString(in.Licy_TxHash)] = preTXOutputs
					}
				}
			}
			// 新增
			for keyHash,outPuts := range outsMap  {
				keyHashBytes,_ := hex.DecodeString(keyHash)
				b.Put(keyHashBytes,outPuts.Licy_Serialize())
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
