//
// machine.go
//
// Distributed under terms of the MIT license.
//

package main

import "C"
import (
	"context"
	"flag"
	"fmt"
	"github.com/AstraProtocol/channel/app"
	channelTypes "github.com/AstraProtocol/channel/x/channel/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	machine "github.com/dungtt-astra/paymentchannel/machine"
	util "github.com/dungtt-astra/paymentchannel/utils"
	"github.com/dungtt-astra/astra-go-sdk/account"
	//"github.com/dungtt-astra/paymentchannel/utils/channel"
	"github.com/dungtt-astra/paymentchannel/utils/config"
	data "github.com/dungtt-astra/paymentchannel/utils/data"
	"github.com/dungtt-astra/paymentchannel/utils/field"
	"github.com/evmos/ethermint/encoding"
	ethermintTypes "github.com/evmos/ethermint/types"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/dungtt-astra/astra-go-sdk/channel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"io"
	"log"
	"time"
)

var (
	serverAddr = flag.String("server_addr", "localhost:9111", "The server address in the format of host:port")
)

type MachineClient struct {
	//machine.UnimplementedMachineClient
	stream    machine.Machine_ExecuteClient
	account   *account.PrivateKeySerialized
	denom     string
	amount    int64
	version   string
	acc_name  string
	channel   data.Msg_Channel
	cn        *channel.Channel
	rpcClient client.Context
}

var mmemonic = "baby cancel magnet patient urge regular ribbon scorpion buyer zoo muffin style echo flock soda text door multiply present vocal budget employ target radar"

var waitc = make(chan struct{})

var cfg = &config.Config{
	ChainId:       "astra_11110-1",
	Endpoint:      "http://128.199.238.171:26657",
	CoinType:      60,
	PrefixAddress: "astra",
	TokenSymbol:   "aastra",
}

func (c *MachineClient) Init(stream machine.Machine_ExecuteClient) {
	c.stream = stream
	c.acc_name = "Alice"
	c.denom = "aastra"
	c.amount = 0
	c.version = "0.1"

	// set astra address
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

	acc, err := account.NewAccount(60).ImportAccount(mmemonic)
	if err != nil {
		log.Println("ImportAccount Err:", err.Error())
		return
	}

	c.account = acc

	c.channel = data.Msg_Channel{
		Index:         "",
		Multisig_Addr: "",
		PartA:         "",
		PartB:         c.account.AccAddress().String(),
		Denom:         c.denom,
		Amount_partA:  0,
		Amount_partB:  float64(c.amount),
		PubkeyA:       c.account.PublicKey().String(),
		PubkeyB:       "",
	}

	//
	rpcClient := client.Context{}
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)

	rpcHttp, err := client.NewClientFromNode(cfg.Endpoint)
	if err != nil {
		panic(err)
	}

	c.rpcClient = rpcClient.
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

	c.cn = channel.NewChannel(c.rpcClient)
}

func connect(serverAddr *string) (machine.Machine_ExecuteClient, *grpc.ClientConn, error) {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	client := machine.NewMachineClient(conn)
	//ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	ctx := context.Background()
	//defer cancel()

	stream, err := client.Execute(ctx)
	if err != nil {
		log.Fatalf("%v.Execute(ctx) = %v, %v: ", client, stream, err)
	}

	return stream, conn, err
}

func (c *MachineClient) openChannel() error {

	var reqOpenMsg = data.Msg_ReqOpen{
		Version:        c.version,
		Account_Name:   c.acc_name,
		Publickey:      c.account.PublicKey().String(),
		Deposit_Amount: 0,
		Deposit_Denom:  c.denom,
		Hashcode:       "abcd",
		MinCoin:        1,
	}

	var item1 = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			field.Version: &structpb.Value{Kind: &structpb.Value_StringValue{
				reqOpenMsg.Version}},
			field.Account_name: &structpb.Value{Kind: &structpb.Value_StringValue{
				reqOpenMsg.Account_Name}},
			field.Public_key: &structpb.Value{Kind: &structpb.Value_StringValue{
				reqOpenMsg.Publickey}},
			field.Deposit_denom: &structpb.Value{Kind: &structpb.Value_StringValue{
				reqOpenMsg.Deposit_Denom}},
			field.Deposit_amount: &structpb.Value{Kind: &structpb.Value_NumberValue{
				reqOpenMsg.Deposit_Amount}},
			field.Hashcode: &structpb.Value{Kind: &structpb.Value_StringValue{
				reqOpenMsg.Hashcode}},
		},
	}

	instruct := machine.Instruction{Cmd: "REQ_OPENCHANNEL", Data: item1}

	log.Println("ReqOpenChannel")
	log.Println("Account Addr:", c.channel.PartB)
	if err := c.stream.Send(&instruct); err != nil {
		log.Fatalf("%v.Send(%v) = %v: ", c.stream, instruct, err)
		return err
	}

	return nil
}

func (s *MachineClient) doConfirmOpenChannel(strSig string) error {

	//hashcode := "abc"
	var item = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"sig_str": &structpb.Value{Kind: &structpb.Value_StringValue{
				strSig}},
		},
	}

	instruct := machine.Instruction{Cmd: "CONFIRM_OPENCHANNEL", Data: item}

	log.Println("RepOpenChannel")
	if err := s.stream.Send(&instruct); err != nil {
		log.Fatalf("%v.Send(%v) = %v: ", s.stream, instruct, err)
		return err
	}
	return nil
}

func (c *MachineClient) handleReplyOpenChannel(data *structpb.Struct) error {

	c.channel.PartA = data.Fields[field.PartA_addr].GetStringValue()
	c.channel.Amount_partA = data.Fields[field.Deposit_amount].GetNumberValue()
	c.channel.PubkeyA = data.Fields[field.Public_key].GetStringValue()
	c.channel.PartB = c.account.AccAddress().String()

	partA_Account, err := account.NewPKAccount(c.channel.PubkeyA)
	if err != nil {
		log.Println("NewPKAccount Err:", err.Error())
		return err
	}

	multisigAddr, multiSigPubKey, err := account.NewAccount(60).CreateMulSignAccountFromTwoAccount(partA_Account.PublicKey(), c.account.PublicKey(), 2)
	if err != nil {
		//c.sendError(err)
		return err
	}
	c.channel.Multisig_Addr = multisigAddr

	log.Println("multisigAddr:", c.channel.Multisig_Addr)

	msg := channelTypes.NewMsgOpenChannel(
		c.channel.Multisig_Addr,
		c.channel.PartA,
		c.channel.PartB,
		&sdk.Coin{
			Denom:  c.channel.Denom,
			Amount: sdk.NewInt(int64(c.channel.Amount_partA)),
		},
		&sdk.Coin{
			Denom:  c.channel.Denom,
			Amount: sdk.NewInt(c.amount),
		},
		c.channel.Multisig_Addr,
	)

	openChannelRequest := channel.SignMsgRequest{
		Msg:      msg,
		GasLimit: 200000,
		GasPrice: "25aastra",
	}

	_, strSig, err := c.cn.SignMultisigMsg(openChannelRequest, c.account, multiSigPubKey)
	if err != nil {
		log.Printf("SignMultisigTxFromOneAccount error: %v\n", err)
		return err
	}

	log.Println("openChannelRequest:", openChannelRequest)
	log.Printf("strSig: %v\n", strSig)

	c.doConfirmOpenChannel(strSig)
	return nil
}

func eventHandler(c *MachineClient) {
	for {
		msg, err := c.stream.Recv()
		if err == io.EOF {
			log.Println("EOF")
			close(waitc)
			return
		}
		if err != nil {
			log.Printf("Err: %v", err)
			close(waitc)
			return
		}
		log.Printf("output: %v", msg.String())

		cmd := msg.GetCmd()
		data := msg.GetData()
		cmd_type := util.CmdType(cmd)

		switch cmd_type {
		case util.REP_OPENCHANNEL:
			log.Println("reply openchanfieldsg")
			log.Println("partA_pubkey:", data.Fields[field.Public_key].GetStringValue())
			c.handleReplyOpenChannel(data)

		case util.MSG_ERROR:
			log.Fatalf("Error: %v", data.Fields[field.Error].GetStringValue())

		default:
			close(waitc)
			//return status.Errorf(codes.Unimplemented, "Operation '%s' not implemented yet", operator)
			log.Println(codes.Unimplemented, "Operation '%s' not implemented yet", cmd)

			return
		}
	}
}

func main() {

	stream, conn, err := connect(serverAddr)

	if err != nil {
		log.Printf("Err: %v", err)
		return
	}

	client := new(MachineClient)

	client.Init(stream)

	go eventHandler(client)

	client.openChannel()

	time.Sleep(500 * time.Millisecond)

	if err := stream.CloseSend(); err != nil {
		log.Fatalf("%v.CloseSend() got error %v, want %v", stream, err, nil)
	}

	<-waitc
	conn.Close()
}
