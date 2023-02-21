package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	signingTypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/ethereum/go-ethereum/accounts"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

func DecodePublicKey(rpcClient client.Context, pkJSON string) (cryptoTypes.PubKey, error) {
	var pk cryptoTypes.PubKey
	err := rpcClient.Codec.UnmarshalInterfaceJSON([]byte(pkJSON), &pk)
	if err != nil {
		return nil, errors.Wrap(err, "UnmarshalInterfaceJSON`")
	}
	return pk, nil
}

func IsMulSign(pk cryptoTypes.PubKey) (bool, error) {
	lpk, ok := pk.(*multisig.LegacyAminoPubKey)
	if !ok {
		return false, nil
	}

	if lpk.Threshold > 1 {
		return true, nil
	}

	return false, nil
}

func IsTxSigner(user types.AccAddress, signers []types.AccAddress) bool {
	for _, s := range signers {
		if bytes.Equal(user.Bytes(), s.Bytes()) {
			return true
		}
	}

	return false
}

func TxBuilderJsonDecoder(txConfig client.TxConfig, txJSON string) ([]byte, error) {
	tx, err := txConfig.TxJSONDecoder()([]byte(txJSON))
	if err != nil {
		return nil, err
	}

	//convert to []byte
	txBytes, err := txConfig.TxEncoder()(tx)
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

func TxBuilderJsonEncoder(txConfig client.TxConfig, tx client.TxBuilder) (string, error) {
	txJSONEncoder, err := txConfig.TxJSONEncoder()(tx.GetTx())
	if err != nil {
		return "", err
	}

	return string(txJSONEncoder), nil
}

func TxBuilderSignatureJsonEncoder(txConfig client.TxConfig, tx client.TxBuilder) (string, error) {
	sigs, err := tx.GetTx().GetSignaturesV2()
	if err != nil {
		return "", err
	}

	json, err := txConfig.MarshalSignatureJSON(sigs)

	return string(json), nil
}

func TxBuilderSignatureJsonDecoder(txConfig client.TxConfig, txJson string) ([]signingTypes.SignatureV2, error) {
	return txConfig.UnmarshalSignatureJSON([]byte(txJson))
}

func SignatureJsonDecoder(txConfig client.TxConfig, txJson string) ([]signingTypes.SignatureV2, error) {
	return txConfig.UnmarshalSignatureJSON([]byte(txJson))
}

func TxHash(txBytes []byte) string {
	return fmt.Sprintf("%X", tmhash.Sum(txBytes))
}

func IsAddressValid(address string) (bool, error) {
	receiver, err := types.AccAddressFromBech32(address)
	if err != nil {
		return false, err
	}

	return receiver.String() == address, nil
}

func EthAddressToCosmosAddress(ethAddress string) (string, error) {
	ethAddr := ethCommon.HexToAddress(ethAddress)
	baseAddr := types.AccAddress(ethAddr.Bytes())
	return baseAddr.String(), nil
}

func CosmosAddressToEthAddress(cosmosAddress string) (string, error) {
	baseAddr, err := types.AccAddressFromBech32(cosmosAddress)
	if err != nil {
		return "", err
	}

	ethAddress := ethCommon.BytesToAddress(baseAddr.Bytes())
	return ethAddress.String(), nil
}

func BlockedStatus(code uint32) bool {
	if code == 0 {
		return true
	}

	return false
}

func VerifyHdPath(hdPath string) (bool, error) {
	_, err := accounts.ParseDerivationPath(hdPath)
	if err != nil {
		return false, errors.Wrap(err, "ParseDerivationPath")
	}

	return true, nil
}

func ConvertToDecimal(amount string, decimal int) (float64, error) {
	if decimal <= 0 {
		decimal = 18
	}

	valFloat, ok := new(big.Float).SetString(amount)
	if !ok {
		return 0, errors.New("can not parser")
	}

	if valFloat.Cmp(big.NewFloat(0)) <= 0 {
		return 0, nil
	}

	coin := big.NewFloat(math.Pow10(int(decimal)))
	result := new(big.Float).Quo(valFloat, coin)

	convert, err := strconv.ParseFloat(result.String(), 64)
	if err != nil {
		return 0, err
	}

	return convert, nil
}

func ToHashCode(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	hashEncode := base64.StdEncoding.EncodeToString(hash[:])
	return hashEncode
}
