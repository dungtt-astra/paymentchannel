//
// machine.go
//
// Distributed under terms of the MIT license.
//

package main

import (
	"context"
	"flag"
	machine "github.com/dungtt-astra/paymentchannel/machine"
	util "github.com/dungtt-astra/paymentchannel/utils"
	data "github.com/dungtt-astra/paymentchannel/utils/data"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"io"
	"log"
	"time"
)

var (
	serverAddr = flag.String("server_addr", "localhost:9111", "The server address in the format of host:port")
)

var waitc = make(chan struct{})

func connect() (machine.Machine_ExecuteClient, *grpc.ClientConn, error) {
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

func openChannel(stream machine.Machine_ExecuteClient, reqOpenMsg data.Msg_ReqOpen) error {

	var item1 = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"version": &structpb.Value{Kind: &structpb.Value_StringValue{
				reqOpenMsg.Version}},
			"account_name": &structpb.Value{Kind: &structpb.Value_StringValue{
				reqOpenMsg.Account_Name}},
			"publickey": &structpb.Value{Kind: &structpb.Value_StringValue{
				reqOpenMsg.Publickey}},
			"deposit_amount": &structpb.Value{Kind: &structpb.Value_NumberValue{
				reqOpenMsg.Deposit_Amount}},
			"deposit_denom": &structpb.Value{Kind: &structpb.Value_StringValue{
				reqOpenMsg.Deposit_Denom}},
			"hashcode": &structpb.Value{Kind: &structpb.Value_StringValue{
				reqOpenMsg.Hashcode}},
		},
	}

	instruct := machine.Instruction{Cmd: "REQ_OPENCHANNEL", Data: item1}

	log.Println("ReqOpenChannel")
	if err := stream.Send(&instruct); err != nil {
		log.Fatalf("%v.Send(%v) = %v: ", stream, instruct, err)
		return err
	}

	return nil
}

func messageHandler(stream machine.Machine_ExecuteClient) {
	for {
		msg, err := stream.Recv()
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
			log.Println("reply openchannel msg")
			log.Println("partner_addr:", data.Fields["account_addr"].GetStringValue())
			log.Println("strSig:", data.Fields["sig_str"].GetStringValue())

		case util.POP:
		default:
			close(waitc)
			//return status.Errorf(codes.Unimplemented, "Operation '%s' not implemented yet", operator)
			log.Println(codes.Unimplemented, "Operation '%s' not implemented yet", cmd)

			return
		}

	}
}

func main() {

	stream, conn, err := connect()

	if err != nil {
		log.Printf("Err: %v", err)
		return
	}

	go messageHandler(stream)

	var req_open_msg = data.Msg_ReqOpen{
		Version:        "0.1",
		Account_Name:   "",
		Publickey:      "string", // account address
		Deposit_Amount: 5,        // sdk.Coin {denom: string, amount: Int}
		Deposit_Denom:  "astra",
		Hashcode:       "abcd",
		MinCoin:        1,
	}
	openChannel(stream, req_open_msg)

	time.Sleep(500 * time.Millisecond)

	if err := stream.CloseSend(); err != nil {
		log.Fatalf("%v.CloseSend() got error %v, want %v", stream, err, nil)
	}

	<-waitc
	conn.Close()
}
