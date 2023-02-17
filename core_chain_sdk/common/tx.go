package common

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/m25-lab/lightning-network-node/core_chain_sdk/account"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authSigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/pkg/errors"
)

type Tx struct {
	txf        tx.Factory
	privateKey *account.PrivateKeySerialized
	rpcClient  client.Context
}

func NewTx(rpcClient client.Context, privateKey *account.PrivateKeySerialized, gasLimit uint64, gasPrice string) *Tx {
	txf := tx.Factory{}.
		WithChainID(rpcClient.ChainID).
		WithTxConfig(rpcClient.TxConfig).
		WithGasPrices(gasPrice).
		WithGas(gasLimit).
		WithSignMode(rpcClient.TxConfig.SignModeHandler().DefaultMode())
	//.SetTimeoutHeight(txf.TimeoutHeight())

	return &Tx{txf: txf, privateKey: privateKey, rpcClient: rpcClient}
}

func (t *Tx) BuildUnsignedTx(msgs types.Msg) (client.TxBuilder, error) {
	return t.txf.BuildUnsignedTx(msgs)
}

func (t *Tx) PrintUnsignedTx(msg types.Msg) (string, error) {
	unsignedTx, err := t.BuildUnsignedTx(msg)
	if err != nil {
		return "", errors.Wrap(err, "BuildUnsignedTx")
	}

	json, err := t.rpcClient.TxConfig.TxJSONEncoder()(unsignedTx.GetTx())
	if err != nil {
		return "", errors.Wrap(err, "TxJSONEncoder")
	}

	return string(json), nil
}

func (t *Tx) prepareSignTx() error {
	//coinType := t.privateKey.CoinType()
	from := t.privateKey.AccAddress()

	//@Todo: tmp
	if err := t.rpcClient.AccountRetriever.EnsureExists(t.rpcClient, from); err != nil {
		return errors.Wrap(err, "EnsureExists")
	}

	initNum, initSeq := t.txf.AccountNumber(), t.txf.Sequence()
	if initNum == 0 || initSeq == 0 {
		var accNum, accSeq uint64
		var err error

		//if coinType == 60 {
		//	hexAddress := common.BytesToAddress(t.privateKey.PublicKey().Address().Bytes())
		//
		//	queryClient := emvTypes.NewQueryClient(t.rpcClient)
		//	cosmosAccount, err := queryClient.CosmosAccount(context.Background(), &emvTypes.QueryCosmosAccountRequest{Address: hexAddress.String()})
		//	if err != nil {
		//		return errors.Wrap(err, "CosmosAccount")
		//	}
		//
		//	accNum = cosmosAccount.AccountNumber
		//	accSeq = cosmosAccount.Sequence
		//
		//} else {
		accNum, accSeq, err = t.rpcClient.AccountRetriever.GetAccountNumberSequence(t.rpcClient, from)
		if err != nil {
			return errors.Wrap(err, "GetAccountNumberSequence")
		}
		//}

		t.txf = t.txf.WithAccountNumber(accNum)
		t.txf = t.txf.WithSequence(accSeq)
	}

	return nil
}

func (t *Tx) SignTx(txBuilder client.TxBuilder) error {
	pubKey := t.privateKey.PublicKey()

	err := t.prepareSignTx()
	if err != nil {
		return errors.Wrap(err, "prepareSignTx")
	}

	sigV2 := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  t.txf.SignMode(),
			Signature: nil,
		},
		Sequence: t.txf.Sequence(),
	}

	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return errors.Wrap(err, "SetSignatures")
	}

	// Construct the SignatureV2 struct
	signerData := authSigning.SignerData{
		ChainID:       t.rpcClient.ChainID,
		AccountNumber: t.txf.AccountNumber(),
		Sequence:      t.txf.Sequence(),
	}

	signWithPrivKey, err := tx.SignWithPrivKey(
		t.txf.SignMode(),
		signerData,
		txBuilder,
		t.privateKey.PrivateKey(),
		t.rpcClient.TxConfig,
		t.txf.Sequence())

	if err != nil {
		return errors.Wrap(err, "SignWithPrivKey")
	}

	err = txBuilder.SetSignatures(signWithPrivKey)
	if err != nil {
		return errors.Wrap(err, "SetSignatures")
	}

	return nil
}
