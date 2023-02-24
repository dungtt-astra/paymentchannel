//
// machine.go
//
// Distributed under terms of the MIT license.
//

package server

import (
	"errors"
	"fmt"
	"github.com/AstraProtocol/channel/app"
	channelTypes "github.com/AstraProtocol/channel/x/channel/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingTypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/dungtt-astra/astra-go-sdk/account"
	"github.com/dungtt-astra/astra-go-sdk/channel"
	"github.com/dungtt-astra/astra-go-sdk/common"
	machine "github.com/dungtt-astra/paymentchannel/machine"
	util "github.com/dungtt-astra/paymentchannel/utils"
	"github.com/dungtt-astra/paymentchannel/utils/config"
	"github.com/dungtt-astra/paymentchannel/utils/data"
	field "github.com/dungtt-astra/paymentchannel/utils/field"
	"github.com/evmos/ethermint/encoding"
	ethermintTypes "github.com/evmos/ethermint/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	structpb "google.golang.org/protobuf/types/known/structpb"
	"io"
	"log"
)

type MachineServer struct {
	machine.UnimplementedMachineServer
	stream      machine.Machine_ExecuteServer
	thisAccount *account.PrivateKeySerialized
	cn          *channel.Channel
	channelInfo data.Msg_Channel
	rpcClient   client.Context
}

type conn struct {
	stream machine.Machine_ExecuteServer
	ip     string
}

var thiscon [2]conn
var server MachineServer
var mmemonic = "leaf emerge will mix junior smile tortoise mystery scheme chair fancy afraid badge carpet pottery raw vicious hood exile amateur symbol battle oyster action"
var cfg = &config.Config{
	ChainId:       "astra_11110-1",
	Endpoint:      "http://128.199.238.171:26657",
	CoinType:      60,
	PrefixAddress: "astra",
	TokenSymbol:   "aastra",
}

func (s *MachineServer) Init(stream machine.Machine_ExecuteServer) {

	s.stream = stream

	sdkConfig := sdk.GetConfig()
	sdkConfig.SetPurpose(44)
	sdkConfig.SetCoinType(ethermintTypes.Bip44CoinType)

	bech32PrefixAccAddr := fmt.Sprintf("%v", cfg.PrefixAddress)
	bech32PrefixAccPub := fmt.Sprintf("%vpub", cfg.PrefixAddress)
	bech32PrefixValAddr := fmt.Sprintf("%vvaloper", cfg.PrefixAddress)
	bech32PrefixValPub := fmt.Sprintf("%vvaloperpub", cfg.PrefixAddress)
	bech32PrefixConsAddr := fmt.Sprintf("%vvalcons", cfg.PrefixAddress)
	bech32PrefixConsPub := fmt.Sprintf("%vvalconspub", cfg.PrefixAddress)

	sdkConfig.SetBech32PrefixForAccount(bech32PrefixAccAddr, bech32PrefixAccPub)
	sdkConfig.SetBech32PrefixForValidator(bech32PrefixValAddr, bech32PrefixValPub)
	sdkConfig.SetBech32PrefixForConsensusNode(bech32PrefixConsAddr, bech32PrefixConsPub)

	rpcClient := client.Context{}
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)

	rpcHttp, err := client.NewClientFromNode(cfg.Endpoint)
	if err != nil {
		panic(err)
	}

	s.rpcClient = rpcClient.
		WithClient(rpcHttp).
		//WithNodeURI(cfg.Endpoint).
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithChainID(cfg.ChainId).
		WithAccountRetriever(authTypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastSync).
		WithTxConfig(encodingConfig.TxConfig)

	s.cn = channel.NewChannel(s.rpcClient)

	acc, err := account.NewAccount(60).ImportAccount(mmemonic)
	if err != nil {
		log.Println("ImportAccount Err:", err.Error())
		return
	}

	s.thisAccount = acc
	s.channelInfo.PartA = s.thisAccount.AccAddress().String()
	s.channelInfo.Amount_partA = 3
}

// var waitc = make(chan struct{})
func (s *MachineServer) validateReqOpenChannelData(data *structpb.Struct) error {

	if len(data.Fields[field.Deposit_denom].GetStringValue()) == 0 {
		return errors.New("Invalid denom")
	}

	if len(data.Fields[field.Account_addr].GetStringValue()) == 0 {
		return errors.New("empty account address")
	}

	acc_addr := data.Fields[field.Account_addr].GetStringValue()

	// debug
	fmt.Printf("AcceptOpenChannel: version %v, account_addr %v, pubkey %v \n",
		data.Fields[field.Version].GetStringValue(),
		data.Fields[field.Account_addr].GetStringValue(),
		data.Fields[field.Public_key].GetStringValue())

	_, err := sdk.AccAddressFromBech32(acc_addr)
	if err != nil {
		return err
	}

	return nil
}

func (s *MachineServer) sendError(e error) error {
	var item = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			field.Error: &structpb.Value{Kind: &structpb.Value_StringValue{
				e.Error()}},
		},
	}

	msg := machine.Instruction{Cmd: util.MSG_ERROR, Data: item}

	if err := s.stream.Send(&msg); err != nil {
		log.Fatalf("%v.Send(%v) = %v: ", s.stream, msg, err)
		return err
	}

	return nil
}

func (s *MachineServer) doReplyOpenChannel(pubkey string) error {

	hashcode := "abc"
	log.Println("start RepOpenChannel")
	//log.Println("PartA AccAddress:", s.thisAccount.AccAddress().String())
	//log.Println("Multisig AccAddress:", s.channelInfo.Multisig_Addr)
	//log.Println("Part A deposit:", s.channelInfo.Amount_partA)

	var item = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			field.PartA_addr: &structpb.Value{Kind: &structpb.Value_StringValue{
				s.thisAccount.AccAddress().String()}},
			field.Multisig_addr: &structpb.Value{Kind: &structpb.Value_StringValue{
				s.channelInfo.Multisig_Addr}},
			field.Public_key: &structpb.Value{Kind: &structpb.Value_StringValue{
				pubkey}},
			field.Deposit_amount: &structpb.Value{Kind: &structpb.Value_NumberValue{
				s.channelInfo.Amount_partA}},
			field.Hashcode: &structpb.Value{Kind: &structpb.Value_StringValue{
				hashcode}},
		},
	}

	instruct := machine.Instruction{Cmd: "REP_OPENCHANNEL", Data: item}

	log.Println("RepOpenChannel")
	if err := s.stream.Send(&instruct); err != nil {
		log.Fatalf("%v.Send(%v) = %v: ", s.stream, instruct, err)
		return err
	}
	return nil
}

func (s *MachineServer) handleReqOpenChannel(data *structpb.Struct) {
	log.Println("Start handle ReqOpenChannel...")

	if err := s.validateReqOpenChannelData(data); err != nil {
		s.sendError(fmt.Errorf("Err: %v", err.Error()))
		return
	}

	//sDec, _ := b64.StdEncoding.DecodeString(sEnc)
	peerPubkeystr := data.Fields[field.Public_key].GetStringValue()
	//peerPubkey := account.

	//, err := util.Pubkey(peerPubkeystr)
	//if err != nil {
	//	s.sendError(err)
	//	return
	//}

	peerAccAddr := data.Fields[field.Account_addr].GetStringValue()

	//@todo create multi account
	log.Println("PartA addr:", s.thisAccount.AccAddress().String())
	log.Println("PartB pubkey addr:", peerPubkey.String())
	multisigAddr, multiSigPubkey, err := account.NewAccount(60).CreateMulSignAccountFromTwoAccount(s.thisAccount.PublicKey(), peerPubkey, 2)
	if err != nil {
		s.sendError(err)
		return
	}

	s.channelInfo.Denom = data.Fields[field.Deposit_denom].GetStringValue()
	s.channelInfo.Amount_partB = data.Fields[field.Deposit_amount].GetNumberValue()
	s.channelInfo.PartB = peerAccAddr
	s.channelInfo.PubkeyB = peerPubkey
	s.channelInfo.Multisig_Addr = multisigAddr
	s.channelInfo.Multisig_Pubkey = multiSigPubkey

	log.Println("multisigAddr:", multisigAddr)

	//log.Println("strSig: ", strSig)
	s.doReplyOpenChannel(s.thisAccount.PublicKey().String())
}

func BuildAndBroadCastMultisigMsg(client client.Context, multiSigPubkey cryptoTypes.PubKey, sig1, sig2 string, msgRequest channel.SignMsgRequest) (*sdk.TxResponse, error) {
	signList := make([][]signingTypes.SignatureV2, 0)

	signByte1, err := common.TxBuilderSignatureJsonDecoder(client.TxConfig, sig1)
	if err != nil {
		return nil, err
	}
	signList = append(signList, signByte1)

	signByte2, err := common.TxBuilderSignatureJsonDecoder(client.TxConfig, sig2)
	if err != nil {
		return nil, err
	}
	signList = append(signList, signByte2)

	newTx := common.NewTxMulSign(client,
		nil,
		msgRequest.GasLimit,
		msgRequest.GasPrice,
		0,
		2,
	)

	txBuilderMultiSign, err := newTx.BuildUnsignedTx(msgRequest.Msg)
	if err != nil {
		return nil, err
	}

	err = newTx.CreateTxMulSign(txBuilderMultiSign, multiSigPubkey, 60, signList)
	if err != nil {
		return nil, err
	}

	txJson, err := common.TxBuilderJsonEncoder(client.TxConfig, txBuilderMultiSign)
	if err != nil {
		return nil, err
	}

	txByte, err := common.TxBuilderJsonDecoder(client.TxConfig, txJson)
	if err != nil {
		return nil, err
	}

	// txHash := common.TxHash(txByte)

	return client.BroadcastTxCommit(txByte)
}

func (s *MachineServer) handleConfirmOpenChannel(data *structpb.Struct) (*sdk.TxResponse, error) {

	msg := channelTypes.NewMsgOpenChannel(
		s.channelInfo.Multisig_Addr,
		s.channelInfo.PartA,
		s.channelInfo.PartB,
		&sdk.Coin{
			Denom:  s.channelInfo.Denom,
			Amount: sdk.NewInt(int64(s.channelInfo.Amount_partA)),
		},
		&sdk.Coin{
			Denom:  s.channelInfo.Denom,
			Amount: sdk.NewInt(int64(s.channelInfo.Amount_partB)),
		},
		s.channelInfo.Multisig_Addr,
	)

	openChannelRequest := channel.SignMsgRequest{
		Msg:      msg,
		GasLimit: 200000,
		GasPrice: "25aastra",
	}

	log.Println("openChannelRequest:", openChannelRequest)
	//log.Printf("multisigaddress: %v\n", s.channelInfo.Multisig_Addr)
	//peerAccount, err := account.NewPKAccount(s.channelInfo.PubkeyB)
	//if err != nil {
	//	log.Println("NewPKAccount Err:", err.Error())
	//	return nil, err
	//}

	//@todo create multi account
	//log.Println("this publickey:", s.thisAccount.PublicKey(), s.thisAccount.PublicKey().String())
	//log.Println("peer publickey:", peerAccount.PublicKey(), s.thisAccount.PublicKey().String())
	multisig_addr, multiSigPubkey, err := account.NewAccount(60).CreateMulSignAccountFromTwoAccount(s.thisAccount.PublicKey(), nil, 2)
	if err != nil {
		s.sendError(err)
		return nil, err
	}
	log.Println("multisig_addr:", multisig_addr)

	strSig1, err := s.cn.SignMultisigMsg(openChannelRequest, s.thisAccount, multiSigPubkey)
	if err != nil {
		log.Println("SignMultisigTxFromOneAccount Err:", err.Error())
		return nil, err
	}

	//@todo create multi account
	strSig2 := data.Fields["sig_str"].GetStringValue()

	//
	log.Println("strSig1:", strSig1)
	log.Println("strSig2:", strSig2)

	txResponse, err := BuildAndBroadCastMultisigMsg(s.rpcClient, multiSigPubkey, strSig1, strSig2, openChannelRequest)
	if err != nil {
		log.Printf("BuildAndBroadCastMultisigMsg Err: %v", err.Error())
		return nil, err
	}

	log.Printf("TXID:%v, gas used:%v, code:%v \n", txResponse.TxHash, txResponse.GasUsed, txResponse.Code)
	return txResponse, nil
}

func eventHandler(s MachineServer, waitc chan bool) {
	for {
		log.Println("wait for receive instruction")
		instruction, err := s.stream.Recv()
		if err == io.EOF {
			log.Println("EOF")
			close(waitc)
			return
		}
		if err != nil {
			return
		}

		cmd := instruction.GetCmd()
		data := instruction.GetData()
		cmd_type := util.CmdType(cmd)

		fmt.Printf("cmd: %v, data: %v\n", cmd, data)

		switch cmd_type {
		case util.REQ_OPENCHANNEL:
			server.handleReqOpenChannel(data)
			//log.Println("AcceptOpenChannel: name %v, ", data.Fields["name"].GetStringValue())

		case util.MSG_ERROR:
			log.Fatalf("%v.Send(%v) = %v: ", s.stream, cmd, data)

		case util.CONFIRM_OPENCHANNEL:
			server.handleConfirmOpenChannel(data)

		default:
			close(waitc)
			//return status.Errorf(codes.Unimplemented, "Operation '%s' not implemented yet", operator)
			log.Println(codes.Unimplemented, "Operation '%s' not implemented yet", cmd)

			return
		}
	}
}

// Execute runs the set of instructions given.
func (s *MachineServer) Execute(stream machine.Machine_ExecuteServer) error {

	ctx := stream.Context()
	pClient, ok := peer.FromContext(ctx)
	pIP := "Can not get IP, so sorry"
	if ok {
		pIP = pClient.Addr.String()
	}
	log.Println(pIP)
	thiscon[0].stream = stream
	thiscon[0].ip = pIP

	waitc := make(chan bool)

	server.Init(stream)

	go eventHandler(server, waitc)

	<-waitc

	return nil

	//var stack stack.Stack
}
