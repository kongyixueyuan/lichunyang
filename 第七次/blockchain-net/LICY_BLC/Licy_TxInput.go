package LICY_BLC

import "bytes"

type Licy_TxInput struct {
	// 1. 交易的Hash
	Licy_TxHash []byte
	// 2. 存储TXOutput在Vout里面的索引
	Licy_Index int

	Licy_Signature []byte // 数字签名

	Licy_PublicKey []byte // 公钥，钱包里面
}

/**
 *  判断当前的消费是谁的钱
 */
func (licy_txInput *Licy_TxInput) Licy_UnLockRipemd160Hash(licy_ripemd160Hash []byte) bool {
	publicKey := Ripemd160Hash(licy_txInput.Licy_PublicKey)
	return bytes.Compare(publicKey,licy_ripemd160Hash) == 0
}