package BLC

import (
	"github.com/boltdb/bolt"
	"log"
	"fmt"
	"time"
	"math/big"
	"os"
)


// 数据库名字
const dbName  = "blockchain.db"

// 表的名字
const blockTableName  = "blocks"

const newBlockHashName = "newBlockHash"

type Blockchain struct {
	Tip []byte //最新的区块的Hash
	DB  *bolt.DB
}

// 判断数据库是否存在
func DBExists() bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}

	return true
}

// 迭代器
func (blockchain *Blockchain) Iterator() *BlockchainIterator {

	return &BlockchainIterator{blockchain.Tip,blockchain.DB}
}

// 遍历输出所有区块的信息
func (blc *Blockchain) Printchain()  {

	blockchainIterator := blc.Iterator()

	for {
		block := blockchainIterator.Next()

		fmt.Printf("Height：%d\n",block.Height)
		fmt.Printf("PrevBlockHash：%x\n",block.PrevBlockHash)
		fmt.Printf("Data：%s\n",block.Data)
		fmt.Printf("Timestamp：%s\n",time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash：%x\n",block.Hash)
		fmt.Printf("Nonce：%d\n",block.Nonce)

		fmt.Println()

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if big.NewInt(0).Cmp(&hashInt) == 0{
			break;
		}
	}

}

// 增加区块到区块链里面
func (blc *Blockchain) AddBlockToBlockchain(data string)  {
	err := blc.DB.Update(func(tx *bolt.Tx) error{

		//1. 获取表
		table := tx.Bucket([]byte(blockTableName))
		//2. 创建新区块
		if table != nil {

			// 先获取最新区块
			blockBytes := table.Get(blc.Tip)
			// 反序列化
			block := DeserializeBlock(blockBytes)

			//3. 将区块序列化并且存储到数据库中
			newBlock := NewBlock(data,block.Height + 1,block.Hash)
			err := table.Put(newBlock.Hash,newBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}
			//4. 更新数据库里面"l"对应的hash
			err = table.Put([]byte(newBlockHashName),newBlock.Hash)
			if err != nil {
				log.Panic(err)
			}
			//5. 更新blockchain的Tip
			blc.Tip = newBlock.Hash
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

}


//1. 创建带有创世区块的区块链
func CreateBlockchainWithGenesisBlock(data string) *Blockchain {
	// 判断数据库是否存在
	if DBExists() {
		fmt.Println("创世区块已经存在.......")
		os.Exit(1)
	}
	fmt.Println("正在创建创世区块.......")
	// 创建或者打开数据库
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	var blockHash []byte

	err = db.Update(func(tx *bolt.Tx) error{
		//  获取表
		table := tx.Bucket([]byte(blockTableName))
		if table == nil {
			// 创建数据库表
			table,err = tx.CreateBucket([]byte(blockTableName))
			if err != nil {
				log.Panic(err)
			}
		}
		if table != nil {
			// 创建创世区块
			genesisBlock := CreateGenesisBlock(data)
			// 将创世区块存储到表中
			err := table.Put(genesisBlock.Hash,genesisBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}
			// 存储最新的区块的hash
			err = table.Put([]byte(newBlockHashName),genesisBlock.Hash)
			if err != nil {
				log.Panic(err)
			}
			blockHash = genesisBlock.Hash
		}

		return nil
	})

	// 返回区块链对象
	return &Blockchain{blockHash,db}

	//// 创建创世区块
	//genesisBlock := CreateGenesisBlock("Genesis Data.......")
	//// 返回区块链对象
	//return &Blockchain{[]*Block{genesisBlock}}
}


// 返回Blockchain对象
func BlockchainObject() *Blockchain {

	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var tip []byte

	err = db.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))

		if b != nil {
			// 读取最新区块的Hash
			tip = b.Get([]byte(newBlockHashName))

		}
		return nil
	})
	return &Blockchain{tip,db}
}