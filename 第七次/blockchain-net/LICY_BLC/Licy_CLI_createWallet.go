package LICY_BLC

import "fmt"

func (cli *Licy_CLI) licy_createWallet(){
	wallets ,_:=Licy_ReadWallets()

	wallets.Licy_CreateNewWallet()

	fmt.Println(len(wallets.Licy_WalletsMap))
}