package LICY_BLC

import (
	"encoding/binary"
	"log"
	"bytes"
	"encoding/json"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"fmt"
	"encoding/gob"
)

//将int64转成byte数组
func Licy_IntToHex(num int64) []byte  {
	buff := new(bytes.Buffer)
	err := binary.Write(buff,binary.BigEndian,num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

// 字节数组反转
func Licy_ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

// 标准的JSON字符串转数组
func Licy_JSONToArray(jsonString string) []string {
	//json 到 []string
	var sArr []string
	if err := json.Unmarshal([]byte(jsonString), &sArr); err != nil {
		log.Panic(err)
	}
	return sArr
}

/**
 * 公钥byte数组转公钥160hash
 */
func Ripemd160Hash(publicKey []byte) []byte {
	//1. 256
	hash256 := sha256.New()
	hash256.Write(publicKey)
	hash := hash256.Sum(nil)
	//2. 160
	ripemd160 := ripemd160.New()
	ripemd160.Write(hash)
	return ripemd160.Sum(nil)
}

//version 转字节数组
func commandToBytes(command string) []byte {
	var bytes [COMMANDLENGTH]byte
	for i, c := range command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}


//字节数组转version
func bytesToCommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}


// 将结构体序列化成字节数组
func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}