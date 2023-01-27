package grpcutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/grpcutil/proto"
)

func TestGrpcClient(t *testing.T) {
	//StartRpcTcpClient()
	StartRpcTlsClient()
}
func StartRpcTcpClient() error {
	// 创建一个 gRPC channel 和服务器交互
	conn, err := InitGrpcTcpClient("localhost:8080")
	if err != nil {
		fmt.Println("StartRpcTcpClient():Dial fail:", err)
		return err
	}
	defer conn.Close()

	// 创建客户端
	client := proto.NewGreeterClient(conn)

	// 直接调用
	resp1, err := client.SayHello(context.Background(), &proto.HelloRequest{
		Name: "Hello Server 1 !!",
	})

	fmt.Println("StartRpcTcpClient(): resp1:", resp1)

	resp2, err := client.SayHello(context.Background(), &proto.HelloRequest{
		Name: "Hello Server 2 !!",
	})

	fmt.Println("StartRpcTcpClient(): resp2:", resp2)
	return nil
}

func StartRpcTlsClient() error {

	conn, err := InitGrpcTlsClient("grpcserver.test.com:8080",
		`..\cert\catlsroot.cer`, `..\cert\clienttlscrt.cer`, `..\cert\clienttlskey.pem`)
	if err != nil {
		fmt.Println("StartRpcTlsClient(): Dial fail:", err)
		return err
	}
	defer conn.Close()

	client := proto.NewGreeterClient(conn)
	resp1, err := client.SayHello(context.Background(), &proto.HelloRequest{
		Name: "Hello Server 1 !!",
	})
	if err != nil {
		fmt.Println("StartRpcTlsClient(): 1 SayHello fail:", err)
		return err
	}

	fmt.Println("StartRpcTlsClient(): resp1:", resp1)

	resp2, err := client.SayHello(context.Background(), &proto.HelloRequest{
		Name: "Hello Server 2 !!",
	})
	if err != nil {
		fmt.Println("StartRpcTlsClient(): 2 SayHello fail:", err)
		return err
	}

	fmt.Println("StartRpcTlsClient(): resp2:", resp2)
	return nil
}
