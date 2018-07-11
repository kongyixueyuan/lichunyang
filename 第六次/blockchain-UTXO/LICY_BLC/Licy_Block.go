package LICY_BLC

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Licy_Block struct {
	Licy_Height int64				//区块高度
	Licy_PrevBlockHash []byte		//上一个区块hash
	Licy_Txs []*Licy_Transaction		//交易数据
	Licy_Timestamp int64			//区块时间戳
	Licy_Nonce int64				//区块难度值
	Licy_Hash []byte				//区块hash值
}

/**
 * 将block序列化成byte数组
 */
func (block *Licy_Block) Licy_Serialize() []byte  {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(block)
	if err != nil{
		log.Panic("block serialize error")
		log.Panic(err)
	}
	return result.Bytes()
}

/**
 * 将txs 转成byte数组
 */
func (licy_block *Licy_Block) Licy_HashTxs() []byte  {
	//var txHashes [][]byte
	//var txHash [32]byte
	//
	//for _, tx := range licy_block.Licy_Txs {
	//	txHashes = append(txHashes, tx.Licy_TxHash)
	//}
	//txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	//
	//return txHash[:]

	var transactions [][]byte

	for _, tx := range licy_block.Licy_Txs {
		transactions = append(transactions, tx.Licy_Serialize())
	}
	mTree := Licy_NewMerkleTree(transactions)

	return mTree.Licy_RootNode.Licy_Data

}

/**
 * 将byte数组反序列化为block
 */
func Licy_DeserializeBlock(blockBytes []byte) *Licy_Block  {
	var licyBlock Licy_Block
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&licyBlock)
	if err != nil {
		log.Panic("block deserialize error")
		log.Panic(err)
	}
	return &licyBlock


}

/**
 * 创建新的区块
 */
func Licy_NewBlock(licyTxs []*Licy_Transaction,licyHeight int64,licyPrevblockHash []byte) *Licy_Block  {
	//new 一个新的区块
	licyBlock := &Licy_Block{licyHeight,licyPrevblockHash,licyTxs,time.Now().Unix(),0,nil}
	// 调用工作量证明的方法并且返回有效的Hash和Nonce
	pow := Licy_NewProofOfWork(licyBlock)
	// 挖矿验证
	hash,nonce := pow.Run()
	licyBlock.Licy_Hash = hash[:]
	licyBlock.Licy_Nonce = nonce
	return licyBlock
}
/**
 * 生成创世区块
 */
func Licy_CreateGenesisBlock(txs []*Licy_Transaction) *Licy_Block {
	return Licy_NewBlock(txs,1, []byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0})
}