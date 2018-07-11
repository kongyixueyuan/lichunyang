package LICY_BLC

import (
	"fmt"
	"os"
)

func (cli *Licy_CLI) printchain()  {

	if Licy_DBExists() == false {
		fmt.Println("数据不存在.......")
		os.Exit(1)
	}

	blockchain := Licy_GetBlochChainObject()

	defer blockchain.Licy_DB.Close()

	blockchain.Licy_Printchain()

}
