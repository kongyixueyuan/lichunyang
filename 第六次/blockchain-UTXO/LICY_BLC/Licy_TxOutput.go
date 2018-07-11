package LICY_BLC

import "bytes"

type Licy_TxOutput struct {
	Licy_Value int64			//收款金额
	Licy_Ripemd160Hash []byte  //address160hash
}

/**
 *  将base58转成160hash
 */
func (licy_TxOutput *Licy_TxOutput)Licy_Lock(address string)  {
	publicKeyHash := Licy_Base58Decode([]byte(address))
	licy_TxOutput.Licy_Ripemd160Hash = publicKeyHash[1:len(publicKeyHash) - 4]
}

func (licy_TxOutput *Licy_TxOutput) Licy_UnLockPubKeyWithAddress(address string) bool  {
	publicKeyHash := Licy_Base58Decode([]byte(address))
	licy_160Hash := publicKeyHash[1:len(publicKeyHash) - 4]
	return bytes.Compare(licy_TxOutput.Licy_Ripemd160Hash,licy_160Hash) ==0
}

func NewLicy_TxOutput(value int64,address string) *Licy_TxOutput {

	txOutput := &Licy_TxOutput{value,nil}
	// 设置Ripemd160Hash
	txOutput.Licy_Lock(address)

	return txOutput
}