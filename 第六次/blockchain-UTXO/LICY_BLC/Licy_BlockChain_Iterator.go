package LICY_BLC

import (
	"github.com/boltdb/bolt"
	"log"
)

//区块链迭代器对象
type Licy_BlockchainIterator struct {
	Licy_CurrentHash []byte
	Licy_DB  *bolt.DB
}

func (blockchainIterator *Licy_BlockchainIterator) Next() *Licy_Block {

	var block *Licy_Block

	err := blockchainIterator.Licy_DB.View(func(tx *bolt.Tx) error{

		blockTable := tx.Bucket([]byte(blockTableName))

		if blockTable != nil {
			currentBlockBytes := blockTable.Get(blockchainIterator.Licy_CurrentHash)
			//  获取到当前迭代器里面的currentHash所对应的区块
			block = Licy_DeserializeBlock(currentBlockBytes)

			// 更新迭代器里面CurrentHash
			blockchainIterator.Licy_CurrentHash = block.Licy_PrevBlockHash
		}

		return nil
	})

	if err != nil {
		log.Panic("blockchain iterator error")
		log.Panic(err)
	}
	return block

}