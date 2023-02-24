package data

import "github.com/evmos/ethermint/crypto/ethsecp256k1"

type Msg_ReqOpen struct {
	Version        string
	Account_Addr   string
	Publickey      string  //
	Deposit_Amount float64 // sdk.Coin {denom: string, amount: Int}
	Deposit_Denom  string
	Hashcode       string
	MinCoin        uint16 // minimum transfer on this channel
}

func (m *Msg_ReqOpen) IsEmpty() bool {
	return len(m.Account_Addr) == 0
}

type Msg_Channel struct {
	Index           string
	Multisig_Addr   string
	Multisig_Pubkey ethsecp256k1.Pubkey
	PartA           string
	PartB           string
	PubkeyA         ethsecp256k1.Pubkey
	PubkeyB         ethsecp256k1.Pubkey
	Denom           string
	Amount_partA    float64
	Amount_partB    float64
}
