package account

import (
	"encoding/hex"
	util "github.com/dungtt-astra/paymentchannel/utils"
	"github.com/pkg/errors"
	"strings"

	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
)

type PKAccount struct {
	publicKey secp256k1.PubKey
}

func NewPKAccount(pubkey string) (*PKAccount, error) {

	if strings.ContainsAny(pubkey, "{") {
		pubkey = strings.Split(strings.Split(pubkey, "{")[1], "}")[0]
	}

	if len(pubkey) != util.PubKeySize {
		return nil, errors.New("length of pubkey is incorrect")
	}

	key, err := hex.DecodeString(pubkey)
	if err != nil {
		return nil, err
	}
	return &PKAccount{
		publicKey: secp256k1.PubKey{
			Key: key,
		},
	}, nil
}

func (pka *PKAccount) PublicKey() cryptoTypes.PubKey {
	return &pka.publicKey
}

func (pka *PKAccount) AccAddress() types.AccAddress {
	pub := pka.PublicKey()
	addr := types.AccAddress(pub.Address())

	return addr
}
