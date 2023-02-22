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
	"github.com/dungtt-astra/paymentchannel/core_chain_sdk/account"
	"github.com/dungtt-astra/paymentchannel/core_chain_sdk/channel"
	"github.com/dungtt-astra/paymentchannel/core_chain_sdk/common"
	machine "github.com/dungtt-astra/paymentchannel/machine"
	util "github.com/dungtt-astra/paymentchannel/utils"
	"github.com/dungtt-astra/paymentchannel/utils/config"
	"github.com/dungtt-astra/paymentchannel/utils/data"
	field "github.com/dungtt-astra/paymentchannel/utils/field"
	"github.com/evmos/ethermint/encoding"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	structpb "google.golang.org/protobuf/types/known/structpb"
	"io"
	"log"
)

type MachineServer struct {
	machine.UnimplementedMachineServer
	stream      machine.Machine_ExecuteServer
	ThisAccount *account.PrivateKeySerialized
	cn          *channel.Channel
	ChannelInfo data.Msg_Channel
	rpcClient   client.Context
}

type conn struct {
	stream machine.Machine_ExecuteServer
	ip     string
}

var thiscon [2]conn
var server MachineServer
var mmemonic = "antenna guard panda arena drill ankle episode render veteran artist simple clerk seminar math cruise speed vacuum visa hen surround impulse ivory special pet"
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

	bech32PrefixAccAddr := fmt.Sprintf("%v", cfg.PrefixAddress)
	bech32PrefixAccPub := fmt.Sprintf("%vpub", cfg.PrefixAddress)
	sdkConfig.SetBech32PrefixForAccount(bech32PrefixAccAddr, bech32PrefixAccPub)

	s.rpcClient = client.Context{}
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)

	rpcHttp, err := client.NewClientFromNode(cfg.Endpoint)
	if err != nil {
		panic(err)
	}

	s.rpcClient = s.rpcClient.
		WithClient(rpcHttp).
		//WithNodeURI(c.endpoint).
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithChainID(cfg.ChainId).
		WithAccountRetriever(authTypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastSync).
		WithTxConfig(encodingConfig.TxConfig)

	s.cn = channel.NewChannel(s.rpcClient)

	acc, err := account.NewAccount().ImportAccount(mmemonic)
	if err != nil {
		log.Println("ImportAccount Err:", err.Error())
		return
	}

	s.ThisAccount = acc
	s.ChannelInfo.PartA = s.ThisAccount.AccAddress().String()
	s.ChannelInfo.Amount_partA = 10
}

// var waitc = make(chan struct{})
func (s *MachineServer) validateReqOpenChannelData(data *structpb.Struct) error {

	if len(data.Fields[field.Deposit_denom].GetStringValue()) == 0 {
		return errors.New("Invalid denom")
	}

	if len(data.Fields[field.Account_name].GetStringValue()) == 0 {
		return errors.New("empty account name")
	}

	acc, err := account.NewPKAccount(data.Fields[field.Public_key].GetStringValue())
	if err != nil {
		return err
	}

	// debug
	fmt.Printf("AcceptOpenChannel: version %v, account %v, pubkey %v, addr %v \n",
		data.Fields[field.Version].GetStringValue(),
		data.Fields[field.Account_name].GetStringValue(),
		data.Fields[field.Public_key].GetNumberValue(),
		acc.AccAddress().String())

	_, err = sdk.AccAddressFromBech32(acc.AccAddress().String())
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

func (s *MachineServer) doReplyOpenChannel(pubkey string, multisig_pubkey string) error {

	hashcode := "abc"
	log.Println("start RepOpenChannel")
	log.Println("AccAddress:", s.ThisAccount.AccAddress().String())
	log.Println("AccAddress:", s.ChannelInfo.Multisig_Addr)
	log.Println("AccAddress:", s.ChannelInfo.Amount_partA)

	var item = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			field.PartA_addr: &structpb.Value{Kind: &structpb.Value_StringValue{
				s.ThisAccount.AccAddress().String()}},
			field.Multisig_addr: &structpb.Value{Kind: &structpb.Value_StringValue{
				s.ChannelInfo.Multisig_Addr}},
			field.Public_key: &structpb.Value{Kind: &structpb.Value_StringValue{
				pubkey}},
			field.Deposit_amount: &structpb.Value{Kind: &structpb.Value_NumberValue{
				s.ChannelInfo.Amount_partA}},
			field.Hashcode: &structpb.Value{Kind: &structpb.Value_StringValue{
				hashcode}},
			field.Multisig_pubkey: &structpb.Value{Kind: &structpb.Value_StringValue{
				multisig_pubkey}},
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
		s.sendError(err)
		log.Println("Err:", err.Error())
		return
	}

	peerAccount, err := account.NewPKAccount(data.Fields[field.Public_key].GetStringValue())
	if err != nil {
		log.Println("NewPKAccount Err:", err.Error())
		return
	}

	//@todo create multi account
	log.Println("this publickey:", s.ThisAccount.PublicKey(), s.ThisAccount.PublicKey().String())
	log.Println("peer publickey:", peerAccount.PublicKey(), s.ThisAccount.PublicKey().String())
	multisigAddr, MultiSigPubkey, err := account.NewAccount().CreateMulSigAccountFromTwoAccount(s.ThisAccount.PublicKey(), peerAccount.PublicKey(), 2)
	if err != nil {
		s.sendError(err)
		return
	}

	s.ChannelInfo.Denom = data.Fields[field.Deposit_denom].GetStringValue()
	s.ChannelInfo.Amount_partB = data.Fields[field.Deposit_amount].GetNumberValue()
	s.ChannelInfo.PartB = peerAccount.AccAddress().String()
	s.ChannelInfo.Multisig_Addr = multisigAddr
	s.ChannelInfo.Multisig_Pubkey = MultiSigPubkey

	log.Println("multisigAddr:", multisigAddr)

	//log.Println("strSig: ", strSig)
	s.doReplyOpenChannel(s.ThisAccount.PublicKey().String(), fmt.Sprintf("%x", MultiSigPubkey.Bytes()))
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
		2)

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
		s.ChannelInfo.Multisig_Addr,
		s.ChannelInfo.PartA,
		s.ChannelInfo.PartB,
		&sdk.Coin{
			Denom:  s.ChannelInfo.Denom,
			Amount: sdk.NewInt(int64(s.ChannelInfo.Amount_partA)),
		},
		&sdk.Coin{
			Denom:  s.ChannelInfo.Denom,
			Amount: sdk.NewInt(int64(s.ChannelInfo.Amount_partB)),
		},
		s.ChannelInfo.Multisig_Addr,
	)

	openChannelRequest := channel.SignMsgRequest{
		Msg:      msg,
		GasLimit: 200000,
		GasPrice: "1aastra",
	}

	strSig1, err := s.cn.SignMultisigTxFromOneAccount(openChannelRequest, s.ThisAccount, s.ChannelInfo.Multisig_Pubkey)

	strSig2 := data.Fields["sig_str"].GetStringValue()

	if err != nil {
		return nil, err
	}

	txResponse, err := BuildAndBroadCastMultisigMsg(s.rpcClient, s.ChannelInfo.Multisig_Pubkey, strSig1, strSig2, openChannelRequest)
	if err != nil {
		return nil, err
	}

	log.Println("TXID:", txResponse.Info)
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
