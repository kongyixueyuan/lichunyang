package LICY_BLC

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"bytes"
	"crypto/sha256"
)

const licy_version = byte(0x00)

const licy_addressChecksumLen = 4

type Licy_Wallet struct {
	//1. 私钥
	Licy_PrivateKey ecdsa.PrivateKey
	//2. 公钥
	Licy_PublicKey  []byte
}
/**
  * 创建钱包
  */
func NewWallet() *Licy_Wallet {

	privateKey,publicKey := licy_newKeyPair()

	return &Licy_Wallet{privateKey,publicKey}
}
/**
 * 通过私钥产生公钥
 */
func licy_newKeyPair() (ecdsa.PrivateKey,[]byte) {

	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic("new KeyPair error")
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pubKey
}

/**
 * 判断地址有效性
 */
func IsValidForAdress(adress []byte) bool {

	// base58 decode
	version_public_checksumBytes := Licy_Base58Decode(adress)

	checkSumBytes := version_public_checksumBytes[len(version_public_checksumBytes) - licy_addressChecksumLen:]

	version_ripemd160 := version_public_checksumBytes[:len(version_public_checksumBytes) - licy_addressChecksumLen]
	//两次256hash
	checkBytes := CheckSum(version_ripemd160)

	if bytes.Compare(checkSumBytes,checkBytes) == 0 {
		return true
	}
	return false
}

/**
 * 获取钱包里的地址
 */
func (w *Licy_Wallet) Licy_GetAddress() []byte  {

	//1. hash160
	// 20字节
	ripemd160Hash := Ripemd160Hash(w.Licy_PublicKey)

	// 21字节
	version_ripemd160Hash := append([]byte{licy_version},ripemd160Hash...)

	// 两次的256 hash
	checkSumBytes := CheckSum(version_ripemd160Hash)

	//25
	bytes := append(version_ripemd160Hash,checkSumBytes...)

	return Licy_Base58Encode(bytes)
}

/**
 * 2次256 hash
 */
func CheckSum(payload []byte) []byte {

	hash1 := sha256.Sum256(payload)

	hash2 := sha256.Sum256(hash1[:])

	return hash2[:licy_addressChecksumLen]
}