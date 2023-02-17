package bank

import (
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"math/big"
)

type TransferRequest struct {
	PrivateKey string
	Receiver   string
	Amount     *big.Int
	GasLimit   uint64
	GasPrice   string
}

type SignTxWithSignerAddressRequest struct {
	SignerPrivateKey string
	SignerPublicKey  cryptoTypes.PubKey
	Receiver         string
	Amount           *big.Int
	GasLimit         uint64
	GasPrice         string
	AccNum           uint64
	SequeNum         uint64
}

type TransferMultiSignRequest struct {
	MulSignAccPublicKey cryptoTypes.PubKey
	Receiver            string
	Amount              *big.Int
	GasLimit            uint64
	GasPrice            string
	Sigs                [][]signing.SignatureV2
	AccNum              uint64
	SequeNum            uint64
}

type Txs struct {
	Code           uint32
	IsOk           bool
	Type           string
	TxDataType     string
	IsUnNativeCoin bool
	Time           string
	BlockHeight    int64
	TxHash         string
	EthTxHash      string
	Sender         string
	EthSender      string
	Receiver       string
	EthReceiver    string
	Amount         string
	AmountDecimal  string
	TokenSymbol    string
	RawData        string
}
