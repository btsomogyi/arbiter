package main

import (
	"github.com/btsomogyi/arbiter/internal"
	"log"
	"net"

	"github.com/btsomogyi/arbiter/example/arbitrated"
	"github.com/btsomogyi/arbiter/example/examplepb"
	"google.golang.org/grpc"
)

func main() {
	supervisor, err := internal.NewSupervisor()
	if err != nil {
		log.Fatal(err)
	}
	go supervisor.Process()
	defer supervisor.Terminate()

	go exampleServer(supervisor)
}

func exampleServer(s *internal.Supervisor) {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	examplepb.RegisterVersionerServer(grpcServer, arbitrated.NewVersioner(s))
	grpcServer.Serve(lis)
}
