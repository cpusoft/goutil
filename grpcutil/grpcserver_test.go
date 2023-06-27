package grpcutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/grpcutil/proto"
)

type HelloServer struct {
	proto.UnimplementedGreeterServer
}

func (s *HelloServer) SayHello(ctx context.Context, in *proto.HelloRequest) (*proto.HelloReply, error) {
	fmt.Println("receive from client:" + in.Name)
	return &proto.HelloReply{Message: "server reply: hi, " + in.Name}, nil
}

func StartRpcTcpServer() {
	listener, grpcServer, _ := InitGrpcTcpServer(":8080")
	// 注册服务
	proto.RegisterGreeterServer(grpcServer, &HelloServer{})
	StartGrpcTcpServer(listener, grpcServer)
}
func StartRpcTlsServer() {
	listener, grpcServer, _ := InitGrpcTlsServer(":8080", `..\cert\catlsroot.cer`,
		`..\cert\clienttlscrt.cer`,
		`..\cert\clienttlskey.pem`)
	proto.RegisterGreeterServer(grpcServer, &HelloServer{}) // &proto.SayHelloServer{})
	StartGrpcTlsServer(listener, grpcServer)
}
func TestGrpcServer(t *testing.T) {
	//StartRpcTcpServer()
	StartRpcTlsServer()
	select {}
}
