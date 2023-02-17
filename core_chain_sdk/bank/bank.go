package bank

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ethCommon "github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	emvTypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/m25-lab/lightning-network-node/core_chain_sdk/account"
	"github.com/m25-lab/lightning-network-node/core_chain_sdk/common"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Bank struct {
	rpcClient   client.Context
	tokenSymbol string
	coinType    uint32
}

func NewBank(rpcClient client.Context, tokenSymbol string, coinType uint32) *Bank {
	return &Bank{rpcClient: rpcClient, tokenSymbol: tokenSymbol, coinType: coinType}
}

func (b *Bank) Balance(addr string) (*big.Int, error) {
	var header metadata.MD

	bankClient := bankTypes.NewQueryClient(b.rpcClient)
	bankRes, err := bankClient.Balance(
		context.Background(),
		&bankTypes.QueryBalanceRequest{Address: addr, Denom: b.tokenSymbol},
		grpc.Header(&header),
	)

	if err != nil {
		return nil, errors.Wrap(err, "Balance")
	}

	return bankRes.Balance.Amount.BigInt(), nil
}

func (b *Bank) AccountRetriever(addr string) (uint64, uint64, error) {
	addrAcc, err := types.AccAddressFromBech32(addr)
	if err != nil {
		return 0, 0, errors.Wrap(err, "AccAddressFromBech32")
	}

	accNum, accSeq, err := b.rpcClient.AccountRetriever.GetAccountNumberSequence(b.rpcClient, addrAcc)
	if err != nil {
		return 0, 0, errors.Wrap(err, "GetAccountNumberSequence")
	}

	return accNum, accSeq, nil
}

func (b *Bank) CheckTx(txHash string) (*types.TxResponse, error) {
	output, err := authtx.QueryTx(b.rpcClient, txHash)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("no transaction found with hash %s. err = %v", txHash, err.Error())
		}

		return nil, err
	}

	if output.Empty() {
		return nil, fmt.Errorf("no transaction found with hash %s", txHash)
	}

	return output, nil
}

func (b *Bank) TransferRawData(param *TransferRequest) (client.TxBuilder, error) {
	auth := account.NewAccount()
	acc, err := auth.ImportAccount(param.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "ImportAccount")
	}

	receiver, err := types.AccAddressFromBech32(param.Receiver)
	if err != nil {
		return nil, errors.Wrap(err, "AccAddressFromBech32")
	}

	coin := types.NewCoin(b.tokenSymbol, types.NewIntFromBigInt(param.Amount))
	msg := bankTypes.NewMsgSend(
		acc.AccAddress(),
		receiver,
		types.NewCoins(coin),
	)

	newTx := common.NewTx(b.rpcClient, acc, param.GasLimit, param.GasPrice)

	txBuilder, err := newTx.BuildUnsignedTx(msg)
	if err != nil {
		return nil, errors.Wrap(err, "BuildUnsignedTx")
	}

	err = newTx.SignTx(txBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "SignTx")
	}

	return txBuilder, nil
}

func (b *Bank) TransferRawDataAndBroadcast(param *TransferRequest) (*sdk.TxResponse, error) {
	auth := account.NewAccount()
	acc, err := auth.ImportAccount(param.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "ImportAccount")
	}

	receiver, err := types.AccAddressFromBech32(param.Receiver)
	if err != nil {
		return nil, errors.Wrap(err, "AccAddressFromBech32")
	}

	coin := types.NewCoin(b.tokenSymbol, types.NewIntFromBigInt(param.Amount))
	msg := bankTypes.NewMsgSend(
		acc.AccAddress(),
		receiver,
		types.NewCoins(coin),
	)

	newTx := common.NewTx(b.rpcClient, acc, param.GasLimit, param.GasPrice)

	txBuilder, err := newTx.BuildUnsignedTx(msg)
	if err != nil {
		return nil, errors.Wrap(err, "BuildUnsignedTx")
	}

	err = newTx.SignTx(txBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "SignTx")
	}

	txJson, err := common.TxBuilderJsonEncoder(b.rpcClient.TxConfig, txBuilder)
	if err != nil {
		panic(err)
	}

	fmt.Println("rawData", string(txJson))

	txByte, err := common.TxBuilderJsonDecoder(b.rpcClient.TxConfig, txJson)
	if err != nil {
		panic(err)
	}

	txHash := common.TxHash(txByte)
	fmt.Println("txHash", txHash)

	fmt.Println(ethCommon.BytesToHash(txByte).String())

	return b.rpcClient.BroadcastTxCommit(txByte)
}

func (b *Bank) TransferRawDataWithPrivateKey(param *TransferRequest) (client.TxBuilder, error) {
	auth := account.NewAccount()
	acc, err := auth.ImportPrivateKey(param.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "ImportAccount")
	}

	receiver, err := types.AccAddressFromBech32(param.Receiver)
	if err != nil {
		return nil, errors.Wrap(err, "AccAddressFromBech32")
	}

	coin := types.NewCoin(b.tokenSymbol, types.NewIntFromBigInt(param.Amount))
	msg := bankTypes.NewMsgSend(
		acc.AccAddress(),
		receiver,
		types.NewCoins(coin),
	)

	newTx := common.NewTx(b.rpcClient, acc, param.GasLimit, param.GasPrice)

	txBuilder, err := newTx.BuildUnsignedTx(msg)
	if err != nil {
		return nil, errors.Wrap(err, "BuildUnsignedTx")
	}

	err = newTx.SignTx(txBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "SignTx")
	}

	return txBuilder, nil
}

func (b *Bank) SignTxWithSignerAddress(param *SignTxWithSignerAddressRequest) (client.TxBuilder, error) {
	auth := account.NewAccount()
	acc, err := auth.ImportAccount(param.SignerPrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "ImportAccount")
	}

	from := types.AccAddress(param.SignerPublicKey.Address())
	receiver, err := types.AccAddressFromBech32(param.Receiver)
	if err != nil {
		return nil, errors.Wrap(err, "AccAddressFromBech32")
	}

	amount := types.NewCoin(b.tokenSymbol, types.NewIntFromBigInt(param.Amount))
	msg := bankTypes.NewMsgSend(
		from,
		receiver,
		types.NewCoins(amount),
	)

	newTx := common.NewMultisigTxBuilder(b.rpcClient, acc, param.GasLimit, param.GasPrice, param.SequeNum, param.AccNum)

	txBuilder, err := newTx.BuildUnsignedTx(msg)
	if err != nil {
		return nil, errors.Wrap(err, "BuildUnsignedTx")
	}

	err = newTx.SignTxWithSignerAddress(txBuilder, param.SignerPublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "SignTxWithSignerAddress")
	}

	return txBuilder, nil
}

func (b *Bank) TransferMultiSignRawData(param *TransferMultiSignRequest) (client.TxBuilder, error) {
	mulSignAccPublicKey := param.MulSignAccPublicKey

	from := types.AccAddress(mulSignAccPublicKey.Address())
	receiver, err := types.AccAddressFromBech32(param.Receiver)
	if err != nil {
		return nil, errors.Wrap(err, "AccAddressFromBech32")
	}

	amount := types.NewCoin(b.tokenSymbol, types.NewIntFromBigInt(param.Amount))
	msg := bankTypes.NewMsgSend(
		from,
		receiver,
		types.NewCoins(amount),
	)

	newTx := common.NewMultisigTxBuilder(b.rpcClient,
		nil,
		param.GasLimit,
		param.GasPrice,
		param.SequeNum,
		param.AccNum)

	txBuilder, err := newTx.BuildUnsignedTx(msg)
	if err != nil {
		return nil, errors.Wrap(err, "BuildUnsignedTx")
	}

	err = newTx.GenerateMultisig(txBuilder, mulSignAccPublicKey, param.Sigs)
	if err != nil {
		return nil, errors.Wrap(err, "GenerateMultisig")
	}

	return txBuilder, nil
}

func (b *Bank) ParserEthMsg(txs *Txs, msgEth *emvTypes.MsgEthereumTx) error {
	ethTx := msgEth.AsTransaction()

	var txDataType string
	switch ethTx.Type() {
	case ethTypes.LegacyTxType:
		txDataType = "legacy_tx_type"
	case ethTypes.AccessListTxType:
		txDataType = "access_list_tx_type"
	case ethTypes.DynamicFeeTxType:
		txDataType = "dynamic_fee_tx_type"
	default:
		return errors.New("type invalid")
	}

	txs.Type = msgEth.Type()
	txs.TxDataType = txDataType
	txs.EthTxHash = ethTx.Hash().String()

	from, err := msgEth.GetSender(ethTx.ChainId())
	if err != nil {
		return errors.Wrap(err, "GetSender")
	}

	sender, err := common.EthAddressToCosmosAddress(from.String())
	if err != nil {
		return errors.Wrap(err, "CosmosAddressToEthAddress")
	}

	txs.Sender = sender
	txs.EthSender = from.String()

	dataExternal := ethTx.Data()
	if dataExternal != nil {
		//todo: can not parser eth data
		txs.IsUnNativeCoin = true
	} else {
		receiver, err := common.EthAddressToCosmosAddress(ethTx.To().String())
		if err != nil {
			return errors.Wrap(err, "EthAddressToCosmosAddress")
		}

		txs.Receiver = receiver
		txs.EthReceiver = ethTx.To().String()
	}

	amount, err := common.ConvertToDecimal(ethTx.Value().String(), 18)
	if err != nil {
		return errors.Wrap(err, "CosmosAddressToEthAddress")
	}

	txs.AmountDecimal = fmt.Sprintf("%v", amount)
	txs.Amount = ethTx.Value().String()

	txs.TokenSymbol = ""

	return nil
}

func (b *Bank) ParserCosmosMsg(txs *Txs, msgSend *bankTypes.MsgSend) error {
	txs.TxDataType = "cosmos"
	txs.Type = msgSend.Type()
	ethSender, err := common.CosmosAddressToEthAddress(msgSend.FromAddress)
	if err != nil {
		return errors.Wrap(err, "CosmosAddressToEthAddress")
	}

	txs.Sender = msgSend.FromAddress
	txs.EthSender = ethSender

	receiver, err := common.CosmosAddressToEthAddress(msgSend.ToAddress)
	if err != nil {
		return errors.Wrap(err, "CosmosAddressToEthAddress")
	}

	txs.Receiver = msgSend.ToAddress
	txs.EthReceiver = receiver

	txs.Amount = msgSend.Amount[0].Amount.String()

	amount, err := common.ConvertToDecimal(msgSend.Amount[0].Amount.String(), 18)
	if err != nil {
		return errors.Wrap(err, "CosmosAddressToEthAddress")
	}

	txs.AmountDecimal = fmt.Sprintf("%v", amount)
	txs.TokenSymbol = msgSend.Amount[0].Denom

	return nil
}

func (b *Bank) TxDetail(txHash string) (*Txs, error) {
	rs, err := b.CheckTx(txHash)
	if err != nil {
		return nil, err
	}

	tx, err := b.rpcClient.TxConfig.TxDecoder()(rs.Tx.GetValue())
	if err != nil {
		return nil, err
	}

	txBytes, err := b.rpcClient.TxConfig.TxJSONEncoder()(tx)
	if err != nil {
		return nil, err
	}

	msg := tx.GetMsgs()[0]
	txs := &Txs{
		Code:        rs.Code,
		IsOk:        common.BlockedStatus(rs.Code),
		Time:        rs.Timestamp,
		BlockHeight: rs.Height,
		TxHash:      rs.TxHash,
		RawData:     string(txBytes),
	}

	msgEth, ok := msg.(*emvTypes.MsgEthereumTx)
	if ok {
		err := b.ParserEthMsg(txs, msgEth)
		if err != nil {
			panic(err)
		}
	}

	msgBankSend, ok := msg.(*bankTypes.MsgSend)
	if ok {
		err := b.ParserCosmosMsg(txs, msgBankSend)
		if err != nil {
			panic(err)
		}
	}

	return txs, nil
}
