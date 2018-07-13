package LICY_BLC

import "fmt"

func (cli *Licy_CLI) Licy_getBalance(address string)  {


	blockchain := Licy_GetBlochChainObject()
	defer blockchain.Licy_DB.Close()

	utxoSet := &Licy_UTXOSet{blockchain}

	amount := utxoSet.Licy_GetBalance(address)

	fmt.Printf("%s一共有%d个Token\n",address,amount)

}