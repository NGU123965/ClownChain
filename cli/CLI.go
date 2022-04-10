package cli

import (
	"ClownChain/blockchain"
	"ClownChain/utils"
	"flag"
	"fmt"
	"log"
	"os"
)

// CLI 结构
type CLI struct {
	BC *blockchain.BlockChain
}

// PrintUsage 展示用法
func PrintUsage() {
	fmt.Println("Usage:")
	fmt.Printf("\tstartnode -- 启动服务. \n")
	fmt.Printf("\ttest -- 测试. \n")
	fmt.Printf("\tcreatewallet -- 创建钱包. \n")
	fmt.Printf("\taddresslists -- 获取钱包地址列表.\n")
	fmt.Printf("\tcreateblockchain -address address -- 地址.\n")
	//fmt.Printf("\taddblock -data DATA -- 交易数据\n")
	fmt.Printf("\tprintchain -- 输出区块链的信息\n")
	fmt.Printf("\tsend -from FROM -to TO -amount AMOUNT -- 转账\n")
	fmt.Printf("\tgetbalance -address FROM -- 查询余额\n")
}

// IsValidArgs 校验，如果只输入了程序命令，就输出指令用法并且退出程序
func IsValidArgs() {
	if len(os.Args) < 2 {
		PrintUsage() // 打印用法
		os.Exit(1)   // 退出程序
	}
}

// Run 运行函数
func (cli *CLI) Run() {
	// 1. 检测参数数量
	IsValidArgs()
	// 获取环境变量
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Println("NODE_ID is not set...")
		os.Exit(1)
	}
	// 2. 新建命令
	// 添加区块
	//addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	// 启动服务
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)
	// 创建钱包
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	// 获取钱包地址
	getAddressListCmd := flag.NewFlagSet("addresslists", flag.ExitOnError)
	// 打印区块链信息
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	// 创建区块链
	createBlCWithGenesisCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	// 发送交易
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	// 查询余额
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	// 添加测试命令
	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	// 3. 获取命令行参数
	//flagAddBlockArg:=addBlockCmd.String("data","send 100 BTC to everyone","交易数据...")
	flagCreateBlockchainWithAddress := createBlCWithGenesisCmd.String("address", "", "地址...")
	// 转账命令行参数
	flagFromArg := sendCmd.String("from", "", "转账地址...")
	flagToArg := sendCmd.String("to", "", "转账目标地址...")
	flagAmount := sendCmd.String("amount", "", "转账金额...")
	// 查询余额命令行参数
	flagBalanceArg := getBalanceCmd.String("address", "", "查询地址...")
	switch os.Args[1] {
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if nil != err {
			log.Panicf("parse cmd of start node failed! %v\n", err)
		}
	case "test":
		err := testCmd.Parse(os.Args[2:])
		if nil != err {
			log.Panicf("parse cmd of test failed! %v\n", err)
		}
	case "addresslists":
		err := getAddressListCmd.Parse(os.Args[2:])
		if nil != err {
			log.Panicf("parse cmd of get address lists failed! %v\n", err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if nil != err {
			log.Panicf("parse cmd of create wallet failed! %v\n", err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if nil != err {
			log.Panicf("parse cmd of send failed! %v\n", err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if nil != err {
			log.Panicf("parse cmd of printchain failed! %v\n", err)
		}
	case "createblockchain":
		err := createBlCWithGenesisCmd.Parse(os.Args[2:])
		if nil != err {
			log.Panicf("parse cmd of create block chain failed! %v\n", err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if nil != err {
			log.Panicf("get balance failed! %v\n", err)
		}
	default:
		PrintUsage()
		os.Exit(1)
	}
	// 启动服务
	if startNodeCmd.Parsed() {
		cli.startNode(nodeID)
	}
	// 测试
	if testCmd.Parsed() {
		//cli.TestMethod()
		cli.TestResetUTXO()
	}
	// 获取钱包地址
	if getAddressListCmd.Parsed() {
		cli.getAddressLists(nodeID)
	}
	// 创建钱包
	if createWalletCmd.Parsed() {
		cli.CreateWallets(nodeID)
	}
	// 添加余额查询命令
	if getBalanceCmd.Parsed() {
		if *flagBalanceArg == "" {
			fmt.Println("未指定查询地址...")
			PrintUsage()
			os.Exit(1)
		}
		cli.getBalance(*flagBalanceArg, nodeID)
	}
	// 添加转账命令
	if sendCmd.Parsed() {
		if *flagFromArg == "" {
			fmt.Println("源地址不能为空...")
			PrintUsage()
			os.Exit(1)
		}
		if *flagToArg == "" {
			fmt.Println("目标地址不能为空...")
			PrintUsage()
			os.Exit(1)
		}
		if *flagAmount == "" {
			fmt.Println("金额不能为空...")
			PrintUsage()
			os.Exit(1)
		}
		cli.send(utils.JSONToArray(*flagFromArg), utils.JSONToArray(*flagToArg), utils.JSONToArray(*flagAmount), nodeID) // 发送交易
	}
	//// 添加区块命令
	//if addBlockCmd.Parsed() {
	//	if *flagAddBlockArg == "" {
	//		PrintUsage()
	//		os.Exit(1)
	//	}
	//	cli.addBlock([]*Transaction{})
	//}
	// 输出区块链信息命令
	if printChainCmd.Parsed() {
		cli.printchain(nodeID)
	}
	// 创建区块链
	if createBlCWithGenesisCmd.Parsed() {
		if *flagCreateBlockchainWithAddress == "" {
			PrintUsage()
			os.Exit(1)
		}
		cli.createBlockchainWithGenesis(*flagCreateBlockchainWithAddress, nodeID)
	}
}
