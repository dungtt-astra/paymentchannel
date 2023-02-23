module github.com/dungtt-astra/paymentchannel

go 1.14

require (
github.com/AstraProtocol/astra-go-sdk v0.0.2
	github.com/AstraProtocol/channel v0.0.11
	github.com/cosmos/cosmos-sdk v0.45.11
	github.com/cosmos/go-bip39 v1.0.0
	github.com/ethereum/go-ethereum v1.11.1
	github.com/evmos/ethermint v0.19.3
	github.com/golang/protobuf v1.5.2 // github.com/toransahu/grpc-eg-go v0.0.0-20210218054324-95afc5b5d158
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.1
	github.com/tendermint/tendermint v0.34.23
	golang.org/x/crypto v0.1.0
	google.golang.org/grpc v1.53.0
	google.golang.org/protobuf v1.28.2-0.20220831092852-f930b1dc76e8
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
