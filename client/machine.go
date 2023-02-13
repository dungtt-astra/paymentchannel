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
	structpb "github.com/golang/protobuf/ptypes/struct"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
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

func openChannel(stream machine.Machine_ExecuteClient) error {

	var item1 = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"name": &structpb.Value{
				Kind: &structpb.Value_StringValue{
					"Anuj1",
				},
			},
			"age": &structpb.Value{
				Kind: &structpb.Value_NumberValue{
					10,
				},
			},
		},
	}

	instruct := machine.Instruction{Cmd: "OPENCHANNEL", Data: item1}

	log.Println("ReqOpenChannel")
	if err := stream.Send(&instruct); err != nil {
		log.Fatalf("%v.Send(%v) = %v: ", stream, instruct, err)
		return err
	}

	return nil
}

func runExecute(client machine.MachineClient, instructions []*machine.Instruction) {
	log.Printf("Streaming %v", instructions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.Execute(ctx)
	if err != nil {
		log.Fatalf("%v.Execute(ctx) = %v, %v: ", client, stream, err)
	}
	//waitc := make(chan struct{})

	go func() {
		for {
			result, err := stream.Recv()
			if err == io.EOF {
				log.Println("EOF")
				close(waitc)
				return
			}
			if err != nil {
				log.Printf("Err: %v", err)
			}
			log.Printf("output: %v", result.GetOutput())
		}
	}()

	for _, instruction := range instructions {
		if err := stream.Send(instruction); err != nil {
			log.Fatalf("%v.Send(%v) = %v: ", stream, instruction, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err := stream.CloseSend(); err != nil {
		log.Fatalf("%v.CloseSend() got error %v, want %v", stream, err, nil)
	}

	<-waitc
}

func listen(stream machine.Machine_ExecuteClient) {
	for {
		result, err := stream.Recv()
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
		log.Printf("output: %v", result.GetOutput())
	}
}

func main() {

	stream, conn, err := connect()

	if err != nil {
		log.Printf("Err: %v", err)
		return
	}

	go listen(stream)

	openChannel(stream)

	time.Sleep(1000 * time.Millisecond)

	if err := stream.CloseSend(); err != nil {
		log.Fatalf("%v.CloseSend() got error %v, want %v", stream, err, nil)
	}

	<-waitc
	conn.Close()
}
