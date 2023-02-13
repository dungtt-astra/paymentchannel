//
// run_machine_server.go
// Copyright (C) 2020 Toran Sahu <toran.sahu@yahoo.com>
//
// Distributed under terms of the MIT license.
//

package main

import (
	"flag"
	"fmt"

	"log"
	"net"

	machine "github.com/dungtt-astra/paymentchannel/machine"
	"github.com/dungtt-astra/paymentchannel/server"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 9111, "Port on which gRPC server should listen TCP conn.")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	machine.RegisterMachineServer(grpcServer, &server.MachineServer{})
	grpcServer.Serve(lis)
	log.Printf("Initializing gRPC server on port %d", *port)
}
