package BLC

import (
	"io"
	"bytes"
	"log"
	"net"
	"fmt"
)


//COMMAND_VERSION
func sendVersion(toAddress string,bc *Blockchain)  {


	bestHeight := bc.GetBestHeight()

	payload := gobEncode(Version{NODE_VERSION, bestHeight, nodeAddress})

	//version
	request := append(commandToBytes(COMMAND_VERSION), payload...)

	fmt.Println("sendVersion:",toAddress)
	sendData(toAddress,request)


}



//COMMAND_GETBLOCKS
func sendGetBlocks(toAddress string)  {

	payload := gobEncode(GetBlocks{nodeAddress})

	request := append(commandToBytes(COMMAND_GETBLOCKS), payload...)
	fmt.Println("sendGetBlocks:",toAddress)
	sendData(toAddress,request)

}

// 主节点将自己的所有的区块hash发送给钱包节点
//COMMAND_BLOCK
//
func sendInv(toAddress string, kind string, hashes [][]byte) {

	payload := gobEncode(Inv{nodeAddress,kind,hashes})

	request := append(commandToBytes(COMMAND_INV), payload...)

	fmt.Println("sendInv:",toAddress)
	sendData(toAddress,request)

}



func sendGetData(toAddress string, kind string ,blockHash []byte) {

	payload := gobEncode(GetData{nodeAddress,kind,blockHash})

	request := append(commandToBytes(COMMAND_GETDATA), payload...)

	fmt.Println("sendGetData:",toAddress)
	sendData(toAddress,request)
}



func sendBlock(toAddress string, block []byte)  {


	payload := gobEncode(BlockData{nodeAddress,block})

	request := append(commandToBytes(COMMAND_BLOCK), payload...)

	fmt.Println("sendBlock:",toAddress)
	sendData(toAddress,request)

}

func sendTx(toAddress string,tx *Transaction)  {


	payload := gobEncode(Tx{nodeAddress,tx})

	request := append(commandToBytes(COMMAND_TX), payload...)

	fmt.Println("sendTx:",toAddress)
	tx.String()
	sendData(toAddress,request)

}

func sendData(to string,data []byte)  {

	fmt.Println("sendData.to:",to)
	fmt.Println(data)
	conn, err := net.Dial("tcp", to)
	if err != nil {
		panic("error")
	}
	defer conn.Close()

	// 附带要发送的数据
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}