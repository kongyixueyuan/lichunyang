package BLC

type TXInput struct {

	TxHash []byte		//交易的Hash

	Vout int			// 存储TXOutput在Vout里面的索引

	ScriptSig string	// 用户名address
}
