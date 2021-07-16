package main

import (
	"context"
	"fmt"
	pb "grpc_study/proto/hello"
	"net"

	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"

	_ "grpc_study/config"
)

var (
	Address = viper.GetString("GRPC_ADDR")
)

type helloService struct{}

var (
	HelloService = helloService{}
	appid        string
	appkey       string
)

func (helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "无Token认证信息")
	}
	if val, ok := md["appid"]; ok {
		appid = val[0]
	}
	if val, ok := md["appkey"]; ok {
		appkey = val[0]
	}

	if appid != viper.GetString("appid") || appkey != viper.GetString("appkey") {
		return nil, status.Errorf(codes.Unauthenticated, "Token认证信息无效: appid=%s, appkey=%s", appid, appkey)
	}

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
	fmt.Printf("Listen on %s with TLS and Token", Address)
	// 启动服务
	s.Serve(listen)
}
