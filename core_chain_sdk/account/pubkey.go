package account

import (
	"encoding/hex"
	"strings"

	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
)

type PKAccount struct {
	publicKey secp256k1.PubKey
}

func NewPKAccount(pubkey string) *PKAccount {
	pubkey = strings.Split(strings.Split(pubkey, "{")[1], "}")[0]
	key, err := hex.DecodeString(pubkey)
	if err != nil {
		panic(err)
	}
	return &PKAccount{
		publicKey: secp256k1.PubKey{
			Key: key,
		},
	}
}

func (pka *PKAccount) PublicKey() cryptoTypes.PubKey {
	return &pka.publicKey
}

func (pka *PKAccount) AccAddress() types.AccAddress {
	pub := pka.PublicKey()
	addr := types.AccAddress(pub.Address())

	return addr
}
