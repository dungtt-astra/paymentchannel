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
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dungtt-astra/paymentchannel/core_chain_sdk/account"
	"github.com/dungtt-astra/paymentchannel/core_chain_sdk/channel"
	machine "github.com/dungtt-astra/paymentchannel/machine"
	util "github.com/dungtt-astra/paymentchannel/utils"
	"github.com/evmos/ethermint/encoding"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	structpb "google.golang.org/protobuf/types/known/structpb"
	"io"
	"log"
)

type MachineServer struct {
	machine.UnimplementedMachineServer
}

type conn struct {
	stream machine.Machine_ExecuteServer
	ip     string
}

var thiscon [2]conn
var server MachineServer
var mmemonic = "excuse quiz oyster vendor often spray day vanish slice topic pudding crew promote floor shadow best subway slush slender good merit hollow certain repeat"

// var waitc = make(chan struct{})
func (s *MachineServer) validateReqOpenChannelData(data *structpb.Struct) error {
	if len(data.Fields["account_name"].GetStringValue()) == 0 {
		return errors.New("empty account name")
	}
	// debug
	fmt.Printf("AcceptOpenChannel: version %v, account %v, amt %v \n",
		data.Fields["version"].GetStringValue(),
		data.Fields["account_name"].GetStringValue(),
		data.Fields["deposit_amount"].GetNumberValue())

	return nil
}

func (s *MachineServer) sendRepOpenChannel(stream machine.Machine_ExecuteServer, msg *channelTypes.MsgOpenChannel, strSig string) error {

	hashcode := "abc"
	var item = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"account_addr": &structpb.Value{Kind: &structpb.Value_StringValue{
				msg.PartA}},
			"multisig_addr": &structpb.Value{Kind: &structpb.Value_StringValue{
				msg.MultisigAddr}},
			//"publickey": &structpb.Value{Kind: &structpb.Value_StringValue{
			//	reqOpenMsg.Publickey}},
			"deposit_amount": &structpb.Value{Kind: &structpb.Value_NumberValue{
				float64(msg.CoinA.Amount.Int64())}},
			"deposit_denom": &structpb.Value{Kind: &structpb.Value_StringValue{
				msg.CoinA.Denom}},
			"hashcode": &structpb.Value{Kind: &structpb.Value_StringValue{
				hashcode}},
			"sig_str": &structpb.Value{Kind: &structpb.Value_StringValue{
				strSig}},
		},
	}

	instruct := machine.Instruction{Cmd: "REP_OPENCHANNEL", Data: item}

	log.Println("RepOpenChannel")
	if err := stream.Send(&instruct); err != nil {
		log.Fatalf("%v.Send(%v) = %v: ", stream, instruct, err)
		return err
	}
	return nil
}

func (s *MachineServer) handleReqOpenChannel(stream machine.Machine_ExecuteServer, data *structpb.Struct) {

	if err := s.validateReqOpenChannelData(data); err != nil {
		log.Println("Err:", err.Error())
		return
	}

	peerAccount := account.NewPKAccount(data.Fields["publickey"].GetStringValue())
	thisAccount, _ := account.NewAccount().ImportAccount(mmemonic)
	//@todo create multi account
	acc := account.NewAccount()
	multisigAddr, multiSigPubkey, _ := acc.CreateMulSigAccountFromTwoAccount(thisAccount.PublicKey(), peerAccount.PublicKey(), 2)

	msg := channelTypes.NewMsgOpenChannel(
		multisigAddr,
		thisAccount.AccAddress().String(),
		peerAccount.AccAddress().String(),
		&sdk.Coin{
			Denom:  data.Fields["deposit_denom"].GetStringValue(),
			Amount: sdk.NewInt(1000000000000000000),
		},
		&sdk.Coin{
			Denom:  "aastra",
			Amount: sdk.NewInt(1000000000000000000),
		},
		multisigAddr,
		"",
	)

	openChannelRequest := channel.SignMsgRequest{
		Msg:      msg,
		GasLimit: 200000,
		GasPrice: "25aastra",
	}

	//channelClient := client.NewChannelClient()
	rpcClient := client.Context{}
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	rpcClient.WithChainID("astra_11110-1").WithTxConfig(encodingConfig.TxConfig)

	//channelClient := client.NewChannelClient()
	strSig, err := channel.NewChannel(rpcClient).SignMultisigTxFromOneAccount(openChannelRequest, thisAccount, multiSigPubkey)

	if err != nil {
		panic(err)
	}

	log.Println(strSig)
	s.sendRepOpenChannel(stream, msg, strSig)
}

func eventHandler(stream machine.Machine_ExecuteServer, waitc chan bool) {
	for {
		log.Println("wait for receive instruction")
		instruction, err := stream.Recv()
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
			server.handleReqOpenChannel(stream, data)
			//log.Println("AcceptOpenChannel: name %v, ", data.Fields["name"].GetStringValue())

		case util.POP:

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

	go eventHandler(stream, waitc)

	<-waitc

	return nil

	//var stack stack.Stack
}
