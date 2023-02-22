package common

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/dungtt-astra/paymentchannel/core_chain_sdk/account"

	keyMultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authSigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/pkg/errors"
)

type TxMulSign struct {
	txf              tx.Factory
	signerPrivateKey *account.PrivateKeySerialized
	rpcClient        client.Context
}

func NewMultisigTxBuilder(rpcClient client.Context, privateKey *account.PrivateKeySerialized, gasLimit uint64, gasPrice string, sequenNum, accNum uint64) *TxMulSign {
	txf := tx.Factory{}.
		WithChainID(rpcClient.ChainID).
		WithTxConfig(rpcClient.TxConfig).
		WithGasPrices(gasPrice).
		WithGas(gasLimit).
		WithSequence(sequenNum).
		WithAccountNumber(accNum).
		WithSignMode(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	//.SetTimeoutHeight(txf.TimeoutHeight())

	return &TxMulSign{txf: txf, signerPrivateKey: privateKey, rpcClient: rpcClient}
}

func (t *TxMulSign) BuildUnsignedTx(msgs types.Msg) (client.TxBuilder, error) {
	return t.txf.BuildUnsignedTx(msgs)
}

func (t *TxMulSign) PrintUnsignedTx(msgs types.Msg) (string, error) {
	unsignedTx, err := t.BuildUnsignedTx(msgs)
	if err != nil {
		return "", errors.Wrap(err, "BuildUnsignedTx")
	}

	json, err := t.rpcClient.TxConfig.TxJSONEncoder()(unsignedTx.GetTx())
	if err != nil {
		return "", errors.Wrap(err, "TxJSONEncoder")
	}

	return string(json), nil
}

func (t *TxMulSign) prepareSignTx(pubKey cryptoTypes.PubKey) error {
	from := types.AccAddress(pubKey.Address())

	if err := t.rpcClient.AccountRetriever.EnsureExists(t.rpcClient, from); err != nil {
		return errors.Wrap(err, "EnsureExists")
	}

	initNum, initSeq := t.txf.AccountNumber(), t.txf.Sequence()
	if initNum == 0 || initSeq == 0 {
		var accNum, accSeq uint64
		var err error

		accNum, accSeq, err = t.rpcClient.AccountRetriever.GetAccountNumberSequence(t.rpcClient, from)
		if err != nil {
			return errors.Wrap(err, "GetAccountNumberSequence")
		}

		t.txf = t.txf.WithAccountNumber(accNum)
		t.txf = t.txf.WithSequence(accSeq)
	}

	return nil
}

func (t *TxMulSign) SignTxWithSignerAddress(txBuilder client.TxBuilder, multiSignAccPubKey cryptoTypes.PubKey) error {
	accMultiSignAddr := types.AccAddress(multiSignAccPubKey.Address())
	if !IsTxSigner(accMultiSignAddr, txBuilder.GetTx().GetSigners()) {
		return fmt.Errorf("address signer %s invalid", accMultiSignAddr.String())
	}

	err := t.prepareSignTx(multiSignAccPubKey)
	if err != nil {
		return errors.Wrap(err, "prepareSignTx")
	}

	pubKey := t.signerPrivateKey.PublicKey()

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
	fmt.Println("$$$", t.rpcClient.ChainID, t.txf.AccountNumber(), t.txf.Sequence())
	signWithPrivKey, err := tx.SignWithPrivKey(
		t.txf.SignMode(),
		signerData,
		txBuilder,
		t.signerPrivateKey.PrivateKey(),
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

func NewTxMulSign(rpcClient client.Context, privateKey *account.PrivateKeySerialized, gasLimit uint64, gasPrice string) *TxMulSign {
	txf := tx.Factory{}.
		WithChainID(rpcClient.ChainID).
		WithTxConfig(rpcClient.TxConfig).
		WithGasPrices(gasPrice).
		WithGas(gasLimit).
		WithSignMode(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	//.SetTimeoutHeight(txf.TimeoutHeight())

	return &TxMulSign{txf: txf, signerPrivateKey: privateKey, rpcClient: rpcClient}
}

//func NewTxMulSign(rpcClient client.Context, privateKey *account.PrivateKeySerialized, gasLimit uint64, gasPrice string, sequenNum, accNum uint64) *TxMulSign {
//	txf := tx.Factory{}.
//		WithChainID(rpcClient.ChainID).
//		WithTxConfig(rpcClient.TxConfig).
//		WithGasPrices(gasPrice).
//		WithGas(gasLimit).
//		WithSequence(sequenNum).
//		WithAccountNumber(accNum).
//		WithSignMode(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
//	//.SetTimeoutHeight(txf.TimeoutHeight())
//
//	return &TxMulSign{txf: txf, signerPrivateKey: privateKey, rpcClient: rpcClient}
//}

func (t *TxMulSign) GenerateMultisig(txBuilder client.TxBuilder, multiSignAccPubKey cryptoTypes.PubKey, signOfSigner [][]signing.SignatureV2) error {
	err := t.prepareSignTx(multiSignAccPubKey)
	if err != nil {
		return errors.Wrap(err, "prepareSignTx")
	}

	multisigPub, ok := multiSignAccPubKey.(*keyMultisig.LegacyAminoPubKey)
	if !ok {
		return errors.Wrap(errors.New("set type error"), "LegacyAminoPubKey")
	}

	multisigSig := multisig.NewMultisig(len(multisigPub.PubKeys))

	for _, v2s := range signOfSigner {
		signingData := authSigning.SignerData{
			ChainID:       t.txf.ChainID(),
			AccountNumber: t.txf.AccountNumber(),
			Sequence:      t.txf.Sequence(),
		}

		for _, sig := range v2s {
			err = authSigning.VerifySignature(
				sig.PubKey,
				signingData,
				sig.Data,
				t.rpcClient.TxConfig.SignModeHandler(),
				txBuilder.GetTx())

			if err != nil {
				addr := types.AccAddress(sig.PubKey.Address())
				return fmt.Errorf("couldn't verify signature for address %s. error = %v", addr.String(), err.Error())
			}

			if err := multisig.AddSignatureV2(multisigSig, sig, multisigPub.GetPubKeys()); err != nil {
				return errors.Wrap(err, "AddSignatureV2")
			}
		}
	}

	sigV2 := signing.SignatureV2{
		PubKey:   multisigPub,
		Data:     multisigSig,
		Sequence: t.txf.Sequence(),
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return errors.Wrap(err, "SetSignatures")
	}

	return nil
}

func (t *TxMulSign) CreateTxMulSign(txBuilder client.TxBuilder, multiSignAccPubKey cryptoTypes.PubKey, coinType uint32, signOfSigner [][]signing.SignatureV2) error {
	err := t.prepareSignTx(multiSignAccPubKey)
	if err != nil {
		return errors.Wrap(err, "prepareSignTx")
	}

	multisigPub, ok := multiSignAccPubKey.(*keyMultisig.LegacyAminoPubKey)
	if !ok {
		return errors.Wrap(errors.New("set type error"), "LegacyAminoPubKey")
	}

	multisigSig := multisig.NewMultisig(len(multisigPub.PubKeys))

	for _, v2s := range signOfSigner {
		signingData := authSigning.SignerData{
			ChainID:       t.txf.ChainID(),
			AccountNumber: t.txf.AccountNumber(),
			Sequence:      t.txf.Sequence(),
		}

		for _, sig := range v2s {
			err = authSigning.VerifySignature(
				sig.PubKey,
				signingData,
				sig.Data,
				t.rpcClient.TxConfig.SignModeHandler(),
				txBuilder.GetTx())

			if err != nil {
				addr := types.AccAddress(sig.PubKey.Address())
				return fmt.Errorf("couldn't verify signature for address %s. error = %v", addr.String(), err.Error())
			}

			if err := multisig.AddSignatureV2(multisigSig, sig, multisigPub.GetPubKeys()); err != nil {
				return errors.Wrap(err, "AddSignatureV2")
			}
		}
	}

	sigV2 := signing.SignatureV2{
		PubKey:   multisigPub,
		Data:     multisigSig,
		Sequence: t.txf.Sequence(),
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return errors.Wrap(err, "SetSignatures")
	}

	return nil
}
