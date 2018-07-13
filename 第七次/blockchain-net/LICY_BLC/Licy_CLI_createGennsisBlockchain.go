package LICY_BLC
// 创建创世区块
func (cli *Licy_CLI) Licy_createGenesisBlockchain(address string)  {

	blockchain := Licy_CreateBlockchainWithGenesisBlock(address)
	defer blockchain.Licy_DB.Close()

	utxoSet := &Licy_UTXOSet{blockchain}

	utxoSet.Licy_ResetUTXOSet()
}
