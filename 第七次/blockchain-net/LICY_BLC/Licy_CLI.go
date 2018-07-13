package LICY_BLC

import (
	"fmt"
	"os"
	"flag"
	"log"
)

type Licy_CLI struct {}

func licy_printUsage()  {

	fmt.Println("Usage:")

	fmt.Println("\taddresslists -- 输出所有钱包地址.")
	fmt.Println("\tcreatewallet -- 创建钱包.")
	fmt.Println("\tcreateblockchain -address -- 交易数据.")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT -- 交易明细.")
	fmt.Println("\tprintchain -- 输出区块信息.")
	fmt.Println("\tgetbalance -address -- 输出区块信息.")
	fmt.Println("\treset -- 重置.")
	fmt.Println("\tstartnode -miner ADDRESS -- 启动节点服务器，并且指定挖矿奖励的地址.")
}

func isValidArgs()  {
	if len(os.Args) < 2 {
		licy_printUsage()
		os.Exit(1)
	}
}

func (cli *Licy_CLI) Run()  {

	isValidArgs()
	//获取节点ID
	// 设置ID
	// export NODE_ID=8888
	// 读取
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID env. var is not set!\n")
		os.Exit(1)
	}
	fmt.Printf("NODE_ID:%s\n",nodeID)


	resetCmd := flag.NewFlagSet("reset",flag.ExitOnError)
	addresslistsCmd := flag.NewFlagSet("addresslists",flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet",flag.ExitOnError)
	sendBlockCmd := flag.NewFlagSet("send",flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain",flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain",flag.ExitOnError)
	getbalanceCmd := flag.NewFlagSet("getbalance",flag.ExitOnError)

	flagFrom := sendBlockCmd.String("from","","转账源地址......")
	flagTo := sendBlockCmd.String("to","","转账目的地地址......")
	flagAmount := sendBlockCmd.String("amount","","转账金额......")

	flagCreateBlockchainWithAddress := createBlockchainCmd.String("address","","创建创世区块的地址")
	getbalanceWithAdress := getbalanceCmd.String("address","","要查询某一个账号的余额.......")

	startNodeCmd := flag.NewFlagSet("startnode",flag.ExitOnError)
	flagMiner := startNodeCmd.String("miner","","定义挖矿奖励的地址......")


	switch os.Args[1] {
	case "send":
		err := sendBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "reset":
		err := resetCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "addresslists":
		err := addresslistsCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getbalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		licy_printUsage()
		os.Exit(1)
	}

	if sendBlockCmd.Parsed() {
		if *flagFrom == "" || *flagTo == "" || *flagAmount == ""{
			licy_printUsage()
			os.Exit(1)
		}



		from := Licy_JSONToArray(*flagFrom)
		to := Licy_JSONToArray(*flagTo)

		for index,fromAdress := range from {
			if IsValidForAdress([]byte(fromAdress)) == false || IsValidForAdress([]byte(to[index])) == false {
				fmt.Printf("地址无效......")
				licy_printUsage()
				os.Exit(1)
			}
		}

		amount := Licy_JSONToArray(*flagAmount)
		cli.send(from,to,amount)
	}

	if printChainCmd.Parsed() {

		//fmt.Println("输出所有区块的数据........")
		cli.printchain()
	}

	if resetCmd.Parsed() {

		//fmt.Println("测试....")
		cli.Licy_ResetMethod()
	}

	if addresslistsCmd.Parsed() {

		//fmt.Println("输出所有区块的数据........")
		cli.licy_addressLists()
	}


	if createWalletCmd.Parsed() {
		// 创建钱包
		cli.licy_createWallet()
	}

	if createBlockchainCmd.Parsed() {

		if IsValidForAdress([]byte(*flagCreateBlockchainWithAddress)) == false {
			fmt.Println("地址无效....")
			licy_printUsage()
			os.Exit(1)
		}


		cli.Licy_createGenesisBlockchain(*flagCreateBlockchainWithAddress)
	}

	if getbalanceCmd.Parsed() {

		if IsValidForAdress([]byte(*getbalanceWithAdress)) == false {
			fmt.Println("地址无效....")
			licy_printUsage()
			os.Exit(1)
		}
		cli.Licy_getBalance(*getbalanceWithAdress)
	}

	if startNodeCmd.Parsed() {
		cli.startNode(nodeID,*flagMiner)
	}

}