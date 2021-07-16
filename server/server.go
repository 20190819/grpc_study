package main

import (
	"context"
	"fmt"
	pb "grpc_study/proto/hello"
	"net"

	"github.com/spf13/viper"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	"google.golang.org/grpc"

	_ "grpc_study/config"
)

var (
	Address = viper.GetString("GRPC_ADDR")
)

type helloService struct{}

var (
	HelloService = helloService{}
)

func (helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	resp := new(pb.HelloResponse)
	resp.Message = fmt.Sprintf("Hello %s \n ", in.Name)

	return resp, nil
}

func main() {
	listen, err := net.Listen("tcp", Address)
	if err != nil {
		fmt.Println(err)
		return
	}

	// tls 认证
	creads, err := credentials.NewServerTLSFromFile("../keys/server.pem", "../keys/server.key")
	if err != nil {
		grpclog.Fatalf("Failed to generate credentials %v", err)
		return
	}
	// 实例化 grpc Server
	s := grpc.NewServer(grpc.Creds(creads))
	// 注册服务
	pb.RegisterHelloServer(s, HelloService)
	fmt.Printf("Listen on %s with TLS", Address)
	// 启动服务
	err = s.Serve(listen)
}
