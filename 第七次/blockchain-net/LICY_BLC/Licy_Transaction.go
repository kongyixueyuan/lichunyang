package LICY_BLC

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
	"crypto/sha256"
	"encoding/hex"
	"crypto/elliptic"
	"math/big"
	"crypto/ecdsa"
	"crypto/rand"
)

type Licy_Transaction struct {

	//1. 交易hash
	Licy_TxHash []byte

	//2. 输入
	Licy_Vins []*Licy_TxInput

	//3. 输出
	Licy_Vouts []*Licy_TxOutput
}

const minerFee = 10

/**
 * 判断交易是否为coinbase
 */
func (tx *Licy_Transaction) Licy_IsCoinbaseTransaction() bool {

	return len(tx.Licy_Vins[0].Licy_TxHash) == 0 && tx.Licy_Vins[0].Licy_Index == -1
}

/**
 * coinBase 交易
 */
func Licy_NewCoinbaseTransaction(address string) *Licy_Transaction {

	//代表消费
	txInput := &Licy_TxInput{[]byte{},-1,nil,[]byte{}}

	txOutput := NewLicy_TxOutput(minerFee,address)

	txCoinbase := &Licy_Transaction{[]byte{},[]*Licy_TxInput{txInput},[]*Licy_TxOutput{txOutput}}

	//设置hash值
	txCoinbase.Licy_HashTransaction()
	return txCoinbase
}

func (tx *Licy_Transaction) Licy_HashTransaction()  {

	//var result bytes.Buffer
	//
	//encoder := gob.NewEncoder(&result)
	//
	//err := encoder.Encode(tx)
	//if err != nil {
	//	log.Panic(err)
	//}

	transactionBytes := tx.Licy_Serialize()
	resultBytes := bytes.Join([][]byte{Licy_IntToHex(time.Now().Unix()),transactionBytes},[]byte{})
	hash := sha256.Sum256(resultBytes)
	tx.Licy_TxHash = hash[:]
}
/**
 * 将Transaction序列化成byte数组
 */
func (licy_Transaction *Licy_Transaction) Licy_Serialize() []byte  {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(licy_Transaction)
	if err != nil{
		log.Panic("Transaction serialize error")
		log.Panic(err)
	}
	return result.Bytes()
}

/**
 * 将byte数组反序列化为Transaction
 */
func Licy_DeserializeTransaction(transactionBytes []byte) *Licy_Transaction  {
	var licy_Transaction Licy_Transaction
	decoder := gob.NewDecoder(bytes.NewReader(transactionBytes))
	err := decoder.Decode(&licy_Transaction)
	if err != nil {
		log.Panic("wallets deserialize error")
		log.Panic(err)
	}
	return &licy_Transaction
}

func Licy_NewSimpleTransaction(from string,to string,amount int64,utxoSet *Licy_UTXOSet,txs []*Licy_Transaction) *Licy_Transaction {


	wallets,_ := Licy_ReadWallets()
	wallet := wallets.Licy_WalletsMap[from]
	// 通过一个函数，返回
	money,spendableUTXODic := utxoSet.Licy_FindSpendableUTXOS(from,amount,txs)
	//
	//	{hash1:[0],hash2:[2,3]}
	var txIntputs []*Licy_TxInput
	var txOutputs []*Licy_TxOutput
	for txHash,indexArray := range spendableUTXODic  {
		txHashBytes,_ := hex.DecodeString(txHash)
		for _,index := range indexArray  {
			txInput := &Licy_TxInput{txHashBytes,index,nil,wallet.Licy_PublicKey}
			txIntputs = append(txIntputs,txInput)
		}
	}
	// 转账
	txOutput := NewLicy_TxOutput(int64(amount),to)
	txOutputs = append(txOutputs,txOutput)
	// 找零
	txOutput = NewLicy_TxOutput(int64(money) - int64(amount),from)
	txOutputs = append(txOutputs,txOutput)
	tx := &Licy_Transaction{[]byte{},txIntputs,txOutputs}
	//设置hash值
	tx.Licy_HashTransaction()

	//进行签名
	utxoSet.Licy_Blockchain.SignTransaction(tx, wallet.Licy_PrivateKey,txs)
	return tx
}

func (tx *Licy_Transaction) Licy_Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Licy_Transaction) {

	if tx.Licy_IsCoinbaseTransaction() {
		return
	}


	for _, vin := range tx.Licy_Vins {
		if prevTXs[hex.EncodeToString(vin.Licy_TxHash)].Licy_TxHash == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}


	txCopy := tx.Licy_TrimmedCopy()

	for inID, vin := range txCopy.Licy_Vins {
		prevTx := prevTXs[hex.EncodeToString(vin.Licy_TxHash)]
		txCopy.Licy_Vins[inID].Licy_Signature = nil
		txCopy.Licy_Vins[inID].Licy_PublicKey = prevTx.Licy_Vouts[vin.Licy_Index].Licy_Ripemd160Hash
		txCopy.Licy_TxHash = txCopy.Licy_Hash()
		txCopy.Licy_Vins[inID].Licy_PublicKey = nil

		// 签名代码
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.Licy_TxHash)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Licy_Vins[inID].Licy_Signature = signature
	}
}

func (tx *Licy_Transaction) Licy_Hash() []byte {

	txCopy := tx

	txCopy.Licy_TxHash = []byte{}

	hash := sha256.Sum256(txCopy.Licy_Serialize())
	return hash[:]
}

// 拷贝一份新的Transaction用于签名                                    T
func (tx *Licy_Transaction) Licy_TrimmedCopy() Licy_Transaction {
	var inputs []*Licy_TxInput
	var outputs []*Licy_TxOutput

	for _, vin := range tx.Licy_Vins {
		inputs = append(inputs, &Licy_TxInput{vin.Licy_TxHash, vin.Licy_Index, nil, nil})
	}

	for _, vout := range tx.Licy_Vouts {
		outputs = append(outputs, &Licy_TxOutput{vout.Licy_Value, vout.Licy_Ripemd160Hash})
	}

	txCopy := Licy_Transaction{tx.Licy_TxHash, inputs, outputs}

	return txCopy
}


// 数字签名验证

func (tx *Licy_Transaction) Licy_Verify(prevTXs map[string]Licy_Transaction) bool {
	if tx.Licy_IsCoinbaseTransaction() {
		return true
	}

	for _, vin := range tx.Licy_Vins {
		if prevTXs[hex.EncodeToString(vin.Licy_TxHash)].Licy_TxHash == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.Licy_TrimmedCopy()

	curve := elliptic.P256()

	for inID, vin := range tx.Licy_Vins {
		prevTx := prevTXs[hex.EncodeToString(vin.Licy_TxHash)]
		txCopy.Licy_Vins[inID].Licy_Signature = nil
		txCopy.Licy_Vins[inID].Licy_PublicKey = prevTx.Licy_Vouts[vin.Licy_Index].Licy_Ripemd160Hash
		txCopy.Licy_TxHash = txCopy.Licy_Hash()
		txCopy.Licy_Vins[inID].Licy_PublicKey = nil


		// 私钥 ID
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Licy_Signature)
		r.SetBytes(vin.Licy_Signature[:(sigLen / 2)])
		s.SetBytes(vin.Licy_Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.Licy_PublicKey)
		x.SetBytes(vin.Licy_PublicKey[:(keyLen / 2)])
		y.SetBytes(vin.Licy_PublicKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.Licy_TxHash, &r, &s) == false {
			return false
		}
	}

	return true
}