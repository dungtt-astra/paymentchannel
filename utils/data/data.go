package data

type Msg_ReqOpen struct {
	Version        string
	Account_Name   string
	Publickey      string  //
	Deposit_Amount float64 // sdk.Coin {denom: string, amount: Int}
	Deposit_Denom  string
	Hashcode       string
	MinCoin        uint16 // minimum transfer on this channel
}

func (m *Msg_ReqOpen) IsEmpty() bool {
	return len(m.Address) == 0
}
