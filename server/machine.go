//
// machine.go
//
// Distributed under terms of the MIT license.
//

package server

import (
	"errors"
	"fmt"
	"github.com/dungtt-astra/paymentchannel/core_chain_sdk/account"
	"github.com/dungtt-astra/paymentchannel/core_chain_sdk/channel"
	machine "github.com/dungtt-astra/paymentchannel/machine"
	"google.golang.org/grpc/codes"
	"io"
	"log"

	channelTypes "github.com/AstraProtocol/channel/x/channel/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/peer"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

type CmdType string

const (
	REQ_OPENCHANNEL CmdType = "REQ_OPENCHANNEL"
	POP                     = "POP"
	ADD                     = "ADD"
	SUB                     = "SUB"
	MUL                     = "MUL"
	DIV                     = "DIV"
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
var this_pubkey = "abc"

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

func (s *MachineServer) sendRepOpenChannel(stream machine.Machine_ExecuteServer, msg sdk.Msg, strSig string) error {
	//stream.
	return nil
}

func (s *MachineServer) handleReqOpenChannel(stream machine.Machine_ExecuteServer, data *structpb.Struct) {

	if err := s.validateReqOpenChannelData(data); err != nil {
		log.Println("Err:", err.Error())
		return
	}

	peerAccount := account.NewPKAccount(data.Fields["publickey"].GetStringValue())
	thisAccount := account.NewPKAccount(this_pubkey)

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
	)

	openChannelRequest := channel.SignMsgRequest{
		Msg:      msg,
		GasLimit: 200000,
		GasPrice: "25aastra",
	}

	//channelClient := client.NewChannelClient()

	strSig, err := channel.SignMultisigTxFromOneAccount(openChannelRequest, thisAccount, multiSigPubkey)
	if err != nil {
		panic(err)
	}

	log.Println(strSig)
	//s.sendRepOpenChannel(stream, Msg, strSig)
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
		cmd_type := CmdType(cmd)

		fmt.Printf("cmd: %v, data: %v\n", cmd, data)

		switch cmd_type {
		case REQ_OPENCHANNEL:
			server.handleReqOpenChannel(stream, data)
			//log.Println("AcceptOpenChannel: name %v, ", data.Fields["name"].GetStringValue())

		case POP:
		case ADD, SUB, MUL, DIV:
			fmt.Printf("!!!")
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
