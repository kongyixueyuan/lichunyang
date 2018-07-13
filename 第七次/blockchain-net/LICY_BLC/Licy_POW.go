package LICY_BLC

import (
	"math/big"
	"bytes"
	"crypto/sha256"
	"fmt"
)


// 256位Hash里面前面至少要有16个零
//targetBit除4为0的个数
const licy_targetBit  = 20 				//4个0

type Licy_ProofOfWork struct {
	Licy_block *Licy_Block // 当前要验证的区块
	licy_target *big.Int // 大数据存储
}

// 数据拼接，返回字节数组
func (pow *Licy_ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Licy_block.Licy_PrevBlockHash,
			pow.Licy_block.Licy_HashTxs(),
			Licy_IntToHex(pow.Licy_block.Licy_Timestamp),
			Licy_IntToHex(int64(licy_targetBit)),
			Licy_IntToHex(int64(nonce)),
			Licy_IntToHex(int64(pow.Licy_block.Licy_Height)),
		},
		[]byte{},
	)
	return data
}

func (licy_proofOfWork *Licy_ProofOfWork) Run() ([]byte,int64) {


	//1. 将Block的属性拼接成字节数组

	//2. 生成hash

	//3. 判断hash有效性，如果满足条件，跳出循环

	nonce := 0

	var hashInt big.Int // 存储我们新生成的hash
	var hash [32]byte

	for {
		//准备数据
		dataBytes := licy_proofOfWork.prepareData(nonce)
		// 生成hash
		hash = sha256.Sum256(dataBytes)
		fmt.Printf("\r%x",hash)
		//判断hashInt是否小于Block里面的target
		if licy_proofOfWork.licy_target.Cmp(&hashInt) == 1 {
			// 将hash存储到hashInt
			hashInt.SetBytes(hash[:])
			break
		}
		nonce = nonce + 1
	}
	return hash[:],int64(nonce)
}


// 创建新的工作量证明对象
func Licy_NewProofOfWork(block *Licy_Block) *Licy_ProofOfWork  {

	target := big.NewInt(1)

	//2. 左移256 - targetBit

	target = target.Lsh(target,256 - licy_targetBit)

	return &Licy_ProofOfWork{block,target}
}

