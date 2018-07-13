package LICY_BLC

import (
	"bytes"
	"encoding/gob"
	"log"
)

type Licy_TxOutputs struct {
	Licy_UTXOS []*Licy_UTXO
}


// 将区块序列化成字节数组
func (txOutputs *Licy_TxOutputs) Licy_Serialize() []byte {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(txOutputs)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// 反序列化
func Licy_DeserializeTXOutputs(txOutputsBytes []byte) *Licy_TxOutputs {

	var txOutputs Licy_TxOutputs

	decoder := gob.NewDecoder(bytes.NewReader(txOutputsBytes))
	err := decoder.Decode(&txOutputs)
	if err != nil {
		log.Panic(err)
	}

	return &txOutputs
}