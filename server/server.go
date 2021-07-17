package main

import (
	"context"
	"fmt"
	pb "grpc_study/proto/hello"
	"net"
	"net/http"

	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"

	_ "grpc_study/config"

	"golang.org/x/net/trace"
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

func main() {
	// tls 认证
	creds, err := credentials.NewServerTLSFromFile("../keys/server.pem", "../keys/server.key")
	if err != nil {
		grpclog.Fatalf("Failed to generate credentials %v", err)
		return
	}
	var opts []grpc.ServerOption
	opts = append(opts, grpc.Creds(creds))
	opts = append(opts, grpc.UnaryInterceptor(interceptor))

	// 实例化 grpc Server
	s := grpc.NewServer(opts...)
	// 注册服务
	pb.RegisterHelloServer(s, HelloService)
	fmt.Printf("Listen on %s with TLS and Token and Interceptor\n", Address)

	listen, err := net.Listen("tcp", Address)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 开启 trance
	go startTrace()

	// 启动服务
	s.Serve(listen)
}

func (helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {

	resp := new(pb.HelloResponse)
	resp.Message = fmt.Sprintf("Hello %s \n ", in.Name)

	return resp, nil
}

func auth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "无Token认证信息")
	}
	if val, ok := md["appid"]; ok {
		appid = val[0]
	}
	if val, ok := md["appkey"]; ok {
		appkey = val[0]
	}

	if appid != viper.GetString("appid") || appkey != viper.GetString("appkey") {
		return status.Errorf(codes.Unauthenticated, "Token认证信息无效: appid=%s, appkey=%s", appid, appkey)
	}
	return nil
}

func interceptor(ctx context.Context, request interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	err := auth(ctx)
	if err != nil {
		return nil, err
	}
	return handler(ctx, request)
}

func startTrace() {
	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
		return true, true
	}
	go http.ListenAndServe(":50051", nil)
	fmt.Println("Trace listen on 50051")
}
