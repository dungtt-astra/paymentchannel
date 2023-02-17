module github.com/dungtt-astra/paymentchannel

go 1.14

require (
	github.com/AstraProtocol/channel v0.0.10
	github.com/cosmos/cosmos-sdk v0.46.10
	github.com/cosmos/go-bip39 v1.0.0
	github.com/ethereum/go-ethereum v1.11.1
	github.com/evmos/ethermint v0.21.0
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2 // github.com/toransahu/grpc-eg-go v0.0.0-20210218054324-95afc5b5d158
	github.com/pkg/errors v0.9.1
	github.com/tendermint/tendermint v0.35.9
	golang.org/x/net v0.6.0 // indirect
	google.golang.org/genproto v0.0.0-20230209215440-0dfe4f8abfcc // indirect
	google.golang.org/grpc v1.53.0
	google.golang.org/protobuf v1.28.2-0.20220831092852-f930b1dc76e8
)

replace (
	github.com/btcsuite/btcd/btcec => github.com/btcsuite/btcd/btcec/v2 v2.1.3
	github.com/btcsuite/btcutil => github.com/btcsuite/btcd/btcutil v1.1.3
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/tendermint/tendermint => github.com/informalsystems/tendermint v0.34.26
)
