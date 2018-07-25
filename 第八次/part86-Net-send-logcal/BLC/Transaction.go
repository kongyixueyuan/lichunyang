package BLC

import (
	"bytes"
	"log"
	"encoding/gob"
	"crypto/sha256"
	"encoding/hex"
	"crypto/ecdsa"
	"crypto/rand"

	"crypto/elliptic"
	"time"
	"fmt"
	"strings"
	"math/big"
)

// UTXO
type Transaction struct {

	//1. 交易hash
	TxHash []byte

	//2. 输入
	Vins []*TXInput

	//3. 输出
	Vouts []*TXOutput
}

//[]byte{}

// 判断当前的交易是否是Coinbase交易
func (tx *Transaction) IsCoinbaseTransaction() bool {

	return len(tx.Vins[0].TxHash) == 0 && tx.Vins[0].Vout == -1
}



//1. Transaction 创建分两种情况
//1. 创世区块创建时的Transaction
func NewCoinbaseTransaction(address string) *Transaction {

	//代表消费
	txInput := &TXInput{[]byte{},-1,nil,[]byte{}}


	txOutput := NewTXOutput(10,address)

	txCoinbase := &Transaction{[]byte{},[]*TXInput{txInput},[]*TXOutput{txOutput}}

	//设置hash值
	txCoinbase.HashTransaction()


	return txCoinbase
}

func (tx *Transaction) HashTransaction()  {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	resultBytes := bytes.Join([][]byte{IntToHex(time.Now().Unix()),result.Bytes()},[]byte{})

	hash := sha256.Sum256(resultBytes)

	tx.TxHash = hash[:]
}



//2. 转账时产生的Transaction

func NewSimpleTransaction(from string,to string,amount int64,utxoSet *UTXOSet,txs []*Transaction,nodeID string) *Transaction {

	//$ ./bc send -from '["juncheng"]' -to '["zhangqiang"]' -amount '["2"]'
	//	[juncheng]
	//	[zhangqiang]
	//	[2]

	wallets,_ := NewWallets(nodeID)
	wallet := wallets.WalletsMap[from]


	// 通过一个函数，返回
	money,spendableUTXODic := utxoSet.FindSpendableUTXOS(from,amount,txs)
	//
	//	{hash1:[0],hash2:[2,3]}

	var txIntputs []*TXInput
	var txOutputs []*TXOutput

	for txHash,indexArray := range spendableUTXODic  {

		txHashBytes,_ := hex.DecodeString(txHash)
		for _,index := range indexArray  {

			txInput := &TXInput{txHashBytes,index,nil,wallet.PublicKey}
			txIntputs = append(txIntputs,txInput)
		}

	}

	// 转账
	txOutput := NewTXOutput(int64(amount),to)
	txOutputs = append(txOutputs,txOutput)

	// 找零
	txOutput = NewTXOutput(int64(money) - int64(amount),from)
	txOutputs = append(txOutputs,txOutput)

	tx := &Transaction{[]byte{},txIntputs,txOutputs}

	//设置hash值
	tx.HashTransaction()

	//进行签名
	utxoSet.Blockchain.SignTransaction(tx, wallet.PrivateKey,txs)

	return tx

}

func (tx *Transaction) Hash() []byte {

	txCopy := tx

	txCopy.TxHash = []byte{}

	fmt.Println("========txCopy Verify.hash start============")
	txCopy.String()
	fmt.Println("========txCopy Verify.hash end============")


	fmt.Println("========txCopy.Serialize() Verify.hash start============")
	s := txCopy.Serialize()
	fmt.Println(s)
	fmt.Println("========txCopy.Serialize() Verify.hash end============")

	hash := sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

func testSerialize()  {
	in := &TXInput{
		[]byte("0a83785cd29f05fa353345f0b282bd2dc5fa5d8bd17fbc6b073bb6ac48bdfc82"),
		1,
		[]byte{},
		[]byte("d3678a1afa55d21bdc544ea858ff933f26645960"),
	}

	out1 := &TXOutput{
		int64(1),
		[]byte("f9afc82dc25caa2a2082e174595196907dbc154e"),
	}
	out2 := &TXOutput{
		int64(8),
		[]byte("d3678a1afa55d21bdc544ea858ff933f26645960"),
	}
	txCopy2 := Transaction{[]byte{},[]*TXInput{in},[]*TXOutput{out1,out2}}
	fmt.Println("========txCopy2.String() Verify.hash start============")
	txCopy2.String()
	fmt.Println("========txCopy2.Serialize() Verify.hash start============")
	//s2,_ := Encode(txCopy2)
	s2  := txCopy2.Serialize()
	fmt.Println(s2)
	fmt.Println("========txCopy2.Serialize() Verify.hash end============")
}


func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// --------------------
// Encode
// 用gob进行数据编码
//
func Encode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
// -------------------
// Decode
// 用gob进行数据解码
//
func Decode(data []byte, to interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(to)
}

//func (tx *Transaction) Serialize() []byte {
//	jsonByte,err := json.Marshal(tx)
//	if err != nil{
//		//fmt.Println("序列化失败:",err)
//		log.Panic(err)
//	}
//	return jsonByte
//}


func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {

	if tx.IsCoinbaseTransaction() {
		return
	}


	for _, vin := range tx.Vins {
		if prevTXs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}


	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vins {
		prevTx := prevTXs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[inID].Signature = nil
		txCopy.Vins[inID].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vins[inID].Signature = signature
		txCopy.Vins[inID].PublicKey = nil
	}

	//for inID, vin := range txCopy.Vins {
	//	prevTx := prevTXs[hex.EncodeToString(vin.TxHash)]
	//	txCopy.Vins[inID].Signature = nil
	//	txCopy.Vins[inID].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash
	//	txCopy.TxHash = txCopy.Hash()
	//	txCopy.Vins[inID].PublicKey = nil
	//
	//	fmt.Println("========txCopy Sign start============")
	//	txCopy.String()
	//	fmt.Println("========txCopy Sign end============")
	//
	//	// 签名代码
	//	r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.TxHash)
	//	if err != nil {
	//		log.Panic(err)
	//	}
	//	signature := append(r.Bytes(), s.Bytes()...)
	//
	//	tx.Vins[inID].Signature = signature
	//}
}


// 拷贝一份新的Transaction用于签名                                    T
func (tx Transaction) TrimmedCopy() Transaction {
	var inputs []*TXInput
	var outputs []*TXOutput

	for _, vin := range tx.Vins {
		inputs = append(inputs, &TXInput{vin.TxHash, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vouts {
		outputs = append(outputs, &TXOutput{vout.Value, vout.Ripemd160Hash})
	}

	txCopy := Transaction{tx.TxHash, inputs, outputs}

	return txCopy
}


// 数字签名验证

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {

	if tx.IsCoinbaseTransaction() {
		return true
	}

	for _, vin := range tx.Vins {
		if prevTXs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	curve := elliptic.P256()
	fmt.Println("========prevTXs Verify start============")
	for key,_tx := range prevTXs{
		fmt.Println("key:",key)
		_tx.String()
	}
	fmt.Println("========prevTXs Verify end============")

	fmt.Println("========tx Verify start============")
	tx.String()
	fmt.Println("========tx Verify end============")



	fmt.Println("========txCopy Verify1 start============")
	txCopy.String()
	fmt.Println("========txCopy Verify1 end============")


	for inID, vin := range tx.Vins {
		prevTx := prevTXs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[inID].Signature = nil
		txCopy.Vins[inID].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PublicKey)
		x.SetBytes(vin.PublicKey[:(keyLen / 2)])
		y.SetBytes(vin.PublicKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false
		}
		txCopy.Vins[inID].PublicKey = nil
	}

	//for inID, vin := range tx.Vins {
	//	prevTx := prevTXs[hex.EncodeToString(vin.TxHash)]
	//	txCopy.Vins[inID].Signature = nil
	//	txCopy.Vins[inID].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash
	//	fmt.Println("========txCopy Verify2 start============")
	//	txCopy.String()
	//	fmt.Println("========txCopy Verify2 end============")
	//	txCopy.TxHash = txCopy.Hash()
	//	fmt.Println("========txCopy Verify start============")
	//	txCopy.String()
	//	fmt.Println("========txCopy Verify end============")
	//	txCopy.Vins[inID].PublicKey = nil
	//
	//
	//	// 私钥 ID
	//	r := big.Int{}
	//	s := big.Int{}
	//	sigLen := len(vin.Signature)
	//	r.SetBytes(vin.Signature[:(sigLen / 2)])
	//	s.SetBytes(vin.Signature[(sigLen / 2):])
	//
	//	x := big.Int{}
	//	y := big.Int{}
	//	keyLen := len(vin.PublicKey)
	//	x.SetBytes(vin.PublicKey[:(keyLen / 2)])
	//	y.SetBytes(vin.PublicKey[(keyLen / 2):])
	//
	//
	//	rawPubKey := ecdsa.PublicKey{curve, &x, &y}
	//	if ecdsa.Verify(&rawPubKey, txCopy.TxHash, &r, &s) == false {
	//		return false
	//	}
	//}

	return true
}

// 打印交易内容
func (tx Transaction) String()  {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction ID: [%x]", tx.TxHash))

	for i, input := range tx.Vins {

		lines = append(lines, fmt.Sprintf("     Input [%d]:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      [%x]", input.TxHash))
		lines = append(lines, fmt.Sprintf("       Out:       [%d]", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: [%x]", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    [%x]", input.PublicKey))
	}

	for i, output := range tx.Vouts {
		lines = append(lines, fmt.Sprintf("     Output [%d]:", i))
		lines = append(lines, fmt.Sprintf("       Value:  [%d]", output.Value))
		lines = append(lines, fmt.Sprintf("       PubKeyHash: [%x]", output.Ripemd160Hash))
	}
	fmt.Println(strings.Join(lines, "\n"))
}
