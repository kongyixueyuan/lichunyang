package LICY_BLC

import (
	"fmt"
	"os"
)

// 转账
func (cli *Licy_CLI) send(from []string,to []string,amount []string)  {


	if Licy_DBExists() == false {
		fmt.Println("数据不存在.......")
		os.Exit(1)
	}

	blockchain := Licy_GetBlochChainObject()
	defer blockchain.Licy_DB.Close()

	blockchain.MineNewBlock(from,to,amount)

	utxoSet := &Licy_UTXOSet{blockchain}

	//转账成功以后，需要更新一下
	utxoSet.Licy_Update()

}