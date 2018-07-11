package LICY_BLC

import "fmt"

/**
 * 输出所有钱包地址
 */
func (cli *Licy_CLI) licy_addressLists()  {
	fmt.Println("输出所有钱包地址：")
	wallets,_ := Licy_ReadWallets()
	for address,_:=range wallets.Licy_WalletsMap  {
		fmt.Println(address)
	}
}
