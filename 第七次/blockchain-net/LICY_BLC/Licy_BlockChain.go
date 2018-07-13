package LICY_BLC

import (
	"github.com/boltdb/bolt"
	"os"
	"fmt"
	"math/big"
	"time"
	"log"
	"encoding/hex"
	"strconv"
	"crypto/ecdsa"
	"bytes"
)

// 数据库名字
const dbName = "Licy_blockchain_%s.db"

// 表的名字
const blockTableName = "Licy_blocks"

const currentHash  = "Licy_currentHash"

type Licy_Blockchain struct {
	Licy_Tip []byte //最新的区块的Hash
	Licy_DB  *bolt.DB
}
// 判断数据库是否存在
func Licy_DBExists() bool {
	nodeID := os.Getenv("NODE_ID")
	dbName := fmt.Sprintf(dbName,nodeID)
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}
	return true
}

// 判断数据库是否存在
//3000
//blockchain_3000.db
func DBExists(dbName string) bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}

	return true
}
// 迭代器
func (blockchain *Licy_Blockchain) Iterator() *Licy_BlockchainIterator {
	return &Licy_BlockchainIterator{blockchain.Licy_Tip, blockchain.Licy_DB}
}

// 遍历输出所有区块的信息
func (blc *Licy_Blockchain) Licy_Printchain() {
	blockchainIterator := blc.Iterator()
	fmt.Println("------------------------------")
	for {
		block := blockchainIterator.Next()
		fmt.Printf("Height：%d\n", block.Licy_Height)
		fmt.Printf("PrevBlockHash：%x\n", block.Licy_PrevBlockHash)
		fmt.Printf("Timestamp：%s\n", time.Unix(block.Licy_Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash：%x\n", block.Licy_Hash)
		fmt.Printf("Nonce：%d\n", block.Licy_Nonce)
		fmt.Println("Txs:")
		for _, tx := range block.Licy_Txs {
			fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
			fmt.Printf("%x\n", tx.Licy_TxHash)
			fmt.Println("Vins:")
			for _, in := range tx.Licy_Vins {
				fmt.Printf("%x\n", in.Licy_TxHash)
				fmt.Printf("%d\n", in.Licy_Index)
				fmt.Printf("%x\n", in.Licy_PublicKey)
			}
			fmt.Println("Vouts:")
			for _, out := range tx.Licy_Vouts {
				fmt.Printf("%d\n",out.Licy_Value)
				fmt.Printf("%x\n",out.Licy_Ripemd160Hash)
			}
			fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		}
		fmt.Println("------------------------------")
		var hashInt big.Int
		hashInt.SetBytes(block.Licy_PrevBlockHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break;
		}
	}
}
/**
 * 添加新的交易到blockchain
 */
func (licy_blc *Licy_Blockchain)Licy_AddBlockToBlockChain(txs []*Licy_Transaction)  {
	err := licy_blc.Licy_DB.Update(func(tx *bolt.Tx) error {
		//1. 获取表
		blockTable := tx.Bucket([]byte(blockTableName))
		//2. 创建新区块
		if blockTable != nil {
			// ⚠️，先获取最新区块
			blockBytes := blockTable.Get(licy_blc.Licy_Tip)
			// 反序列化
			block := Licy_DeserializeBlock(blockBytes)
			//3. 将区块序列化并且存储到数据库中
			newBlock := Licy_NewBlock(txs, block.Licy_Height+1, block.Licy_Hash)
			err := blockTable.Put(newBlock.Licy_Hash, newBlock.Licy_Serialize())
			if err != nil {
				log.Panic("add block to block chain serialize error ")
				log.Panic(err)
			}
			//4. 更新数据库里面"l"对应的hash
			err = blockTable.Put([]byte(currentHash), newBlock.Licy_Hash)
			if err != nil {
				log.Panic("put currentHash  error ")
				log.Panic(err)
			}
			//5. 更新blockchain的Tip
			licy_blc.Licy_Tip = newBlock.Licy_Hash
		}
		return nil
	})
	if err != nil {
		log.Panic("DB error ")
		log.Panic(err)
	}
}

//1. 创建带有创世区块的区块链
func Licy_CreateBlockchainWithGenesisBlock(address string) *Licy_Blockchain {

	nodeID := os.Getenv("NODE_ID")
	dbName := fmt.Sprintf(dbName,nodeID)

	// 判断数据库是否存在
	if Licy_DBExists() {
		fmt.Println("创世区块已经存在.......")
		os.Exit(1)
	}
	fmt.Println("正在创建创世区块.......")
	// 创建或者打开数据库
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	var genesisHash []byte
	err = db.Update(func(tx *bolt.Tx) error {
		// 创建数据库表
		blockChainTable, err := tx.CreateBucket([]byte(blockTableName))
		if err != nil {
			log.Panic(err)
		}
		if blockChainTable != nil {
			// 创建创世区块
			// 创建了一个coinbase Transaction
			txCoinbase := Licy_NewCoinbaseTransaction(address)
			genesisBlock := Licy_CreateGenesisBlock([]*Licy_Transaction{txCoinbase})
			// 将创世区块存储到表中
			err := blockChainTable.Put(genesisBlock.Licy_Hash, genesisBlock.Licy_Serialize())
			if err != nil {
				log.Panic(err)
			}
			// 存储最新的区块的hash
			err = blockChainTable.Put([]byte(currentHash), genesisBlock.Licy_Hash)
			if err != nil {
				log.Panic(err)
			}
			genesisHash = genesisBlock.Licy_Hash
		}
		return nil
	})
	return &Licy_Blockchain{genesisHash, db}
}

/**
 * 返回blockchain对象
 */
func Licy_GetBlochChainObject() *Licy_Blockchain  {
	nodeID := os.Getenv("NODE_ID")
	dbName := fmt.Sprintf(dbName,nodeID)
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var tip []byte
	err = db.View(func(tx *bolt.Tx) error {
		blockTable := tx.Bucket([]byte(blockTableName))
		if blockTable != nil {
			// 读取最新区块的Hash
			tip = blockTable.Get([]byte(currentHash))
		}
		return nil
	})
	return &Licy_Blockchain{tip, db}
}

// 如果一个地址对应的TXOutput未花费，那么这个Transaction就应该添加到数组中返回
func (blockchain *Licy_Blockchain) Licy_UnUTXOs(address string,txs []*Licy_Transaction) []*Licy_UTXO {

	var unUTXOs []*Licy_UTXO

	spentTXOutputs := make(map[string][]int) //hash : index
	for _,tx := range txs {
		if tx.Licy_IsCoinbaseTransaction() == false {
			for _, in := range tx.Licy_Vins {
				//是否能够解锁
				publicKeyHash := Licy_Base58Decode([]byte(address))
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

			if out.Licy_UnLockPubKeyWithAddress(address) {
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
	blockIterator := blockchain.Iterator()
	for {
		block := blockIterator.Next()
		for i := len(block.Licy_Txs) - 1; i >= 0 ; i-- {
			tx := block.Licy_Txs[i]
			// txHash
			// Vins
			if tx.Licy_IsCoinbaseTransaction() == false {
				for _, in := range tx.Licy_Vins {
					//是否能够解锁
					publicKeyHash := Licy_Base58Decode([]byte(address))
					ripemd160Hash := publicKeyHash[1:len(publicKeyHash) - 4]
					if in.Licy_UnLockRipemd160Hash(ripemd160Hash) {
						key := hex.EncodeToString(in.Licy_TxHash)
						spentTXOutputs[key] = append(spentTXOutputs[key], in.Licy_Index)
					}
				}
			}
			// Vouts
		work:
			for index, out := range tx.Licy_Vouts {
				if out.Licy_UnLockPubKeyWithAddress(address) {
					if spentTXOutputs != nil {
						if len(spentTXOutputs) != 0 {
							var isSpentUTXO bool
							for txHash, indexArray := range spentTXOutputs {
								for _, i := range indexArray {
									if index == i && txHash == hex.EncodeToString(tx.Licy_TxHash) {
										isSpentUTXO = true
										continue work
									}
								}
							}
							if isSpentUTXO == false {
								utxo := &Licy_UTXO{tx.Licy_TxHash, index, out}
								unUTXOs = append(unUTXOs, utxo)
							}
						} else {
							utxo := &Licy_UTXO{tx.Licy_TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}
		}
		var hashInt big.Int
		hashInt.SetBytes(block.Licy_PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break;
		}
	}
	return unUTXOs
}
// 转账时查找可用的UTXO
func (blockchain *Licy_Blockchain) Licy_FindSpendableUTXOS(from string, amount int,txs []*Licy_Transaction) (int64, map[string][]int) {

	//1. 现获取所有的UTXO
	utxos := blockchain.Licy_UnUTXOs(from,txs)
	spendableUTXO := make(map[string][]int)
	//2. 遍历utxos
	var value int64
	for _, utxo := range utxos {
		value = value + utxo.Licy_Output.Licy_Value
		hash := hex.EncodeToString(utxo.Licy_TxHash)
		spendableUTXO[hash] = append(spendableUTXO[hash], utxo.Licy_Index)
		if value >= int64(amount) {
			break
		}
	}
	if value < int64(amount) {
		fmt.Printf("%s's fund is 不足\n", from)
		os.Exit(1)
	}
	return value, spendableUTXO
}
// 挖掘新的区块
func (blockchain *Licy_Blockchain) MineNewBlock(from []string, to []string, amount []string) {

	//1.建立一笔交易

	utxoSet := &Licy_UTXOSet{blockchain}

	var txs []*Licy_Transaction

	for index,address := range from {
		value, _ := strconv.Atoi(amount[index])
		tx := Licy_NewSimpleTransaction(address, to[index], int64(value), utxoSet,txs)
		txs = append(txs, tx)
		//fmt.Println(tx)
	}

	//奖励
	tx := Licy_NewCoinbaseTransaction(from[0])
	txs = append(txs,tx)


	//1. 通过相关算法建立Transaction数组
	var block *Licy_Block

	blockchain.Licy_DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))
		if b != nil {

			hash := b.Get([]byte(currentHash))

			blockBytes := b.Get(hash)

			block = Licy_DeserializeBlock(blockBytes)
		}

		return nil
	})

	// 在建立新区块之前对txs进行签名验证
	_txs := []*Licy_Transaction{}
	for _,tx := range txs  {
		if blockchain.Licy_VerifyTransaction(tx,_txs) != true {
			log.Panic("ERROR: Invalid transaction")
		}
		_txs = append(_txs,tx)
	}

	//2. 建立新的区块
	block = Licy_NewBlock(txs, block.Licy_Height+1, block.Licy_Hash)

	//将新区块存储到数据库
	blockchain.Licy_DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {

			b.Put(block.Licy_Hash, block.Licy_Serialize())

			b.Put([]byte(currentHash), block.Licy_Hash)

			blockchain.Licy_Tip = block.Licy_Hash

		}
		return nil
	})

}

// 查询余额
func (blockchain *Licy_Blockchain) GetBalance(address string) int64 {

	utxos := blockchain.Licy_UnUTXOs(address,[]*Licy_Transaction{})

	var amount int64

	for _, utxo := range utxos {

		amount = amount + utxo.Licy_Output.Licy_Value
	}

	return amount
}

func (bclockchain *Licy_Blockchain) SignTransaction(tx *Licy_Transaction,privKey ecdsa.PrivateKey,txs []*Licy_Transaction)  {

	if tx.Licy_IsCoinbaseTransaction() {
		return
	}

	prevTXs := make(map[string]Licy_Transaction)

	for _, vin := range tx.Licy_Vins {
		prevTX, err := bclockchain.Licy_FindTransaction(vin.Licy_TxHash,txs)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.Licy_TxHash)] = prevTX
	}

	tx.Licy_Sign(privKey, prevTXs)

}


func (bc *Licy_Blockchain) Licy_FindTransaction(ID []byte,txs []*Licy_Transaction) (Licy_Transaction, error) {


	for _,tx := range txs  {
		if bytes.Compare(tx.Licy_TxHash, ID) == 0 {
			return *tx, nil
		}
	}


	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Licy_Txs {
			if bytes.Compare(tx.Licy_TxHash, ID) == 0 {
				return *tx, nil
			}
		}

		var hashInt big.Int
		hashInt.SetBytes(block.Licy_PrevBlockHash)

		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break;
		}
	}
	return Licy_Transaction{},nil
}

// 验证数字签名
func (bc *Licy_Blockchain) Licy_VerifyTransaction(tx *Licy_Transaction,txs []*Licy_Transaction) bool {

	prevTXs := make(map[string]Licy_Transaction)

	for _, vin := range tx.Licy_Vins {
		prevTX, err := bc.Licy_FindTransaction(vin.Licy_TxHash,txs)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.Licy_TxHash)] = prevTX
	}
	return tx.Licy_Verify(prevTXs)
}

// [string]*TXOutputs
func (blc *Licy_Blockchain) Licy_FindUTXOMap() map[string]*Licy_TxOutputs  {

	blcIterator := blc.Iterator()

	// 存储已花费的UTXO的信息
	spentableUTXOsMap := make(map[string][]*Licy_TxInput)

	utxoMaps := make(map[string]*Licy_TxOutputs)

	for {
		block := blcIterator.Next()

		for i := len(block.Licy_Txs) - 1; i >= 0 ;i-- {
			txOutputs := &Licy_TxOutputs{[]*Licy_UTXO{}}
			tx := block.Licy_Txs[i]

			// coinbase
			if tx.Licy_IsCoinbaseTransaction() == false {
				for _,txInput := range tx.Licy_Vins {

					txHash := hex.EncodeToString(txInput.Licy_TxHash)
					spentableUTXOsMap[txHash] = append(spentableUTXOsMap[txHash],txInput)
				}
			}
			txHash := hex.EncodeToString(tx.Licy_TxHash)

		WorkOutLoop:
			for index,out := range tx.Licy_Vouts  {

				if tx.Licy_IsCoinbaseTransaction() {
					//fmt.Println("IsCoinbaseTransaction")
					//fmt.Println(out)
					//fmt.Println(txHash)
				}

				txInputs := spentableUTXOsMap[txHash]
				if len(txInputs) > 0 {
					isSpent := false
					for _,in := range  txInputs {
						outPublicKey := out.Licy_Ripemd160Hash
						inPublicKey := in.Licy_PublicKey
						if bytes.Compare(outPublicKey,Ripemd160Hash(inPublicKey)) == 0{
							if index == in.Licy_Index {
								isSpent = true
								continue WorkOutLoop
							}
						}
					}
					if isSpent == false {
						utxo := &Licy_UTXO{tx.Licy_TxHash,index,out}
						txOutputs.Licy_UTXOS = append(txOutputs.Licy_UTXOS,utxo)
					}
				} else {
					utxo := &Licy_UTXO{tx.Licy_TxHash,index,out}
					txOutputs.Licy_UTXOS = append(txOutputs.Licy_UTXOS,utxo)
				}
			}
			// 设置键值对
			utxoMaps[txHash] = txOutputs
		}
		// 找到创世区块时退出
		var hashInt big.Int
		hashInt.SetBytes(block.Licy_PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break;
		}
	}
	return utxoMaps
}

func (bc *Licy_Blockchain) GetBestHeight() int64 {
	block := bc.Iterator().Next()
	return block.Licy_Height
}


func (bc *Licy_Blockchain) GetBlockHashes() [][]byte {
	blockIterator := bc.Iterator()
	var blockHashs [][]byte
	for {
		block := blockIterator.Next()
		blockHashs = append(blockHashs,block.Licy_Hash)
		var hashInt big.Int
		hashInt.SetBytes(block.Licy_PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break;
		}
	}
	return blockHashs
}

// 返回Blockchain对象
func BlockchainObject(nodeID string) *Licy_Blockchain {
	dbName := fmt.Sprintf(dbName,nodeID)
	// 判断数据库是否存在
	if DBExists(dbName) == false {
		fmt.Println("数据库不存在....")
		os.Exit(1)
	}
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	var tip []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			// 读取最新区块的Hash
			tip = b.Get([]byte(currentHash))
		}
		return nil
	})

	return &Licy_Blockchain{tip, db}
}
func (bc *Licy_Blockchain) Licy_AddBlock(block *Licy_Block)  {
	err := bc.Licy_DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			blockExist := b.Get(block.Licy_Hash)
			if blockExist != nil {
				// 如果存在，不需要做任何过多的处理
				return nil
			}

			err := b.Put(block.Licy_Hash,block.Licy_Serialize())

			if err != nil {
				log.Panic(err)
			}
			// 最新的区块链的Hash
			blockHash := b.Get([]byte(currentHash))
			blockBytes := b.Get(blockHash)
			blockInDB := Licy_DeserializeBlock(blockBytes)
			if blockInDB.Licy_Height < block.Licy_Height {
				b.Put([]byte(currentHash),block.Licy_Hash)
				bc.Licy_Tip = block.Licy_Hash
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}














