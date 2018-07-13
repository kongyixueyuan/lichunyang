package LICY_BLC

func (cli *Licy_CLI) Licy_ResetMethod()  {

	blockchain := Licy_GetBlochChainObject()

	defer blockchain.Licy_DB.Close()

	utxoSet := &Licy_UTXOSet{blockchain}

	utxoSet.Licy_ResetUTXOSet()

	//fmt.Println(blockchain.FindUTXOMap())
}
