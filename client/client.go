package main

import (
	"fmt"
	_ "grpc_study/config"
	pb "grpc_study/proto/hello"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

const (
	SERVER_COMMON_NAME = "www.hello.com"
)

var (
	Address = viper.GetString("GRPC_ADDR")
	OpenTLS = viper.GetBool("OpenTLS")
	appid   = viper.GetString("appid")
	appkey  = viper.GetString("appkey")
)

type customCredentials struct{}

// 自定义认证接口
func (customCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"appid":  appid,
		"appkey": appkey,
	}, nil
}

// 自定义是否开启 tls
func (customCredentials) RequireTransportSecurity() bool {
	return OpenTLS
}

func main() {

	var err error
	var opts []grpc.DialOption

	if OpenTLS {
		// tls 认证
		creads, err := credentials.NewClientTLSFromFile("../keys/server.pem", SERVER_COMMON_NAME)
		if err != nil {
			grpclog.Fatalf("Failed to generate TLS credentials %v", err)
			return
		}
		opts = append(opts, grpc.WithTransportCredentials(creads))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	opts = append(opts, grpc.WithPerRPCCredentials(new(customCredentials)))
	opts = append(opts, grpc.WithUnaryInterceptor(interceptor))

	// 连接
	conn, err := grpc.Dial(Address, opts...)
	if err != nil {
		grpclog.Fatalln("server dial error: ", err)
		return
	}

	defer conn.Close()

	// 初始化客户端
	c := pb.NewHelloClient(conn)
	// 调用方法
	request := &pb.HelloRequest{Name: "yangliang"}
	res, err := c.SayHello(context.Background(), request)

	if err != nil {
		grpclog.Fatalln(err)
		return
	}

	fmt.Println(res.Message)
}

// interceptor 客户端拦截器
func interceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	fmt.Printf("method=%s req=%v rep=%v duration=%s error=%v \n", method, req, reply, time.Since(start), err)
	return err
}
