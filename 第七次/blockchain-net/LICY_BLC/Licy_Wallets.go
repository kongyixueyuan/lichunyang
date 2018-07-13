package LICY_BLC

import (
	"fmt"
	"bytes"
	"encoding/gob"
	"crypto/elliptic"
	"log"
	"io/ioutil"
	"os"
)

const walletFile  = "Wallets.dat"

type Licy_Wallets struct {
	//map key为公钥base58字符串，value为对应钱包对象
	Licy_WalletsMap map[string]*Licy_Wallet
}

// 创建一个新钱包
func (w *Licy_Wallets) Licy_CreateNewWallet()  {
	wallet := NewWallet()
	fmt.Printf("Address：%s\n",wallet.Licy_GetAddress())
	w.Licy_WalletsMap[string(wallet.Licy_GetAddress())] = wallet
	w.Licy_SaveWallets()
}

/**
 * 钱包信息写入文件
 */
func (licy_Wallets Licy_Wallets)Licy_SaveWallets()  {

	walletInfoBytes := licy_Wallets.Licy_Serialize()

	err := ioutil.WriteFile(walletFile,walletInfoBytes,0644)
	if err != nil {
		log.Panic("write wallets error")
		log.Panic(err)
	}
}

/**
 * 读钱包文件，返回钱包s
 * 钱包文件不存在，返回空钱包对象s
 */
func Licy_ReadWallets() (*Licy_Wallets,error)  {
	//判断是否存在。不存在则新建一个空wallet
	if _,err := os.Stat(walletFile);os.IsNotExist(err){
		wallets := &Licy_Wallets{}
		wallets.Licy_WalletsMap = make(map[string]*Licy_Wallet)
		return wallets,err
	}
	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}
	wallets := Licy_DeserializeWallets(fileContent)
	return wallets,nil
}

/**
 * 将Wallets序列化成byte数组
 */
func (licy_Wallets *Licy_Wallets) Licy_Serialize() []byte  {
	var result bytes.Buffer
	// 注册的目的，是为了，可以序列化任何类型
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(licy_Wallets)
	if err != nil{
		log.Panic("wallets serialize error")
		log.Panic(err)
	}
	return result.Bytes()
}

/**
 * 将byte数组反序列化为wallets
 */
func Licy_DeserializeWallets(walletsBytes []byte) *Licy_Wallets  {
	var licy_Wallets Licy_Wallets
	// 注册的目的，是为了，可以序列化任何类型
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(walletsBytes))
	err := decoder.Decode(&licy_Wallets)
	if err != nil {
		log.Panic("wallets deserialize error")
		log.Panic(err)
	}
	return &licy_Wallets
}