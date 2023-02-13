//
// machine.go
//
// Distributed under terms of the MIT license.
//

package server

import (
	"fmt"
	machine "github.com/dungtt-astra/paymentchannel/machine"
	"google.golang.org/grpc/codes"
	"io"
	"log"

	"google.golang.org/grpc/peer"
)

type CmdType string

const (
	OPENCHANNEL CmdType = "OPENCHANNEL"
	POP                 = "POP"
	ADD                 = "ADD"
	SUB                 = "SUB"
	MUL                 = "MUL"
	DIV                 = "DIV"
)

type MachineServer struct {
	machine.UnimplementedMachineServer
}

type conn struct {
	stream machine.Machine_ExecuteServer
	ip     string
}

var thiscon [2]conn

//var waitc = make(chan struct{})

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
		case OPENCHANNEL:
			//log.Println("AcceptOpenChannel: name %v, ", data.Fields["name"].GetStringValue())
			fmt.Printf("AcceptOpenChannel: name %v, age %v \n", data.Fields["name"].GetStringValue(),
				data.Fields["age"].GetNumberValue())

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
