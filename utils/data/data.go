package data

type msg_ReqOpen struct {
	Version        string
	Account_Name   string
	Address        string  // account address
	Deposit_Amount float64 // sdk.Coin {denom: string, amount: Int}
	Deposit_Denom  string
	Hashcode       string
	MinCoin        uint16 // minimum transfer on this channel
}
