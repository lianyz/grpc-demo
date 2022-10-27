/*
@Time : 2022/10/25 21:42
@Author : lianyz
@Description :
*/

package main

import (
	"google.golang.org/grpc"
	"log"
	"net"

	pb "github.com/lianyz/product/csi"
)

const (
	port = ":50051"
)

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed listen: %v", err)
		return
	}

	rpcServer := grpc.NewServer()
	pb.RegisterProductInfoServer(rpcServer, &service{})
	pb.RegisterOrderManagementServer(rpcServer, &service{})

	log.Printf("Sttarting gRPC listener on port " + port)
	if err := rpcServer.Serve(listener); err != nil {
		log.Fatalf("failed tot serve: %v", err)
	}
}
