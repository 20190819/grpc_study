package main

import (
	"fmt"
	_ "grpc_study/config"
	pb "grpc_study/proto/hello"

	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

var (
	Address = viper.GetString("GRPC_ADDR")
)

func main() {
	// tls 认证
	creads, err := credentials.NewClientTLSFromFile("../keys/server.pem", "go-grpc-example")
	if err != nil {
		grpclog.Fatalf("Failed to generate TLS credentials %v", err)
		return
	}
	// 连接
	conn, err := grpc.Dial(Address, grpc.WithTransportCredentials(creads))
	if err != nil {
		grpclog.Fatalln("server dial error: ", err)
		return
	}

	defer conn.Close()

	// 初始化客户端
	c := pb.NewHelloClient(conn)
	// 调用方法
	query := &pb.HelloRequest{Name: "wangwu"}
	res, err := c.SayHello(context.Background(), query)

	if err != nil {
		grpclog.Fatalln(err)
		return
	}

	fmt.Println(res.Message)
}
