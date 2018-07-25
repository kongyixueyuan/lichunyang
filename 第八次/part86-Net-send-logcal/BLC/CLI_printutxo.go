package BLC

func (cli *CLI) printutxo(nodeID string) {
	bc := BlockchainObject(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.DB.Close()
	UTXOSet.String()
}
