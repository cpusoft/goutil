package grpcutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/cpusoft/goutil/grpcutil/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestGrpcClient(t *testing.T) {
	//StartRpcTcpClient()
	StartRpcTlsClient()
}
func StartRpcTcpClient() error {
	// 创建一个 gRPC channel 和服务器交互
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
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
	cert, err := tls.LoadX509KeyPair(`..\cert\clienttlscrt.cer`,
		`..\cert\clienttlskey.pem`)
	if err != nil {
		fmt.Println("StartRpcTlsClient(): LoadX509KeyPair fail:", err)
		return err
	}
	// 将根证书加入证书池
	certPool := x509.NewCertPool()
	bs, err := ioutil.ReadFile(`..\cert\catlsroot.cer`)
	if err != nil {
		fmt.Println("StartRpcTlsClient(): ReadFile fail:", err)
		return err
	}

	if !certPool.AppendCertsFromPEM(bs) {
		fmt.Println("StartRpcTlsClient(): AppendCertsFromPEM fail:")
		return errors.New("append cert fail")
	}

	// 新建凭证
	// ServerName 需要与服务器证书内的通用名称一致
	transportCreds := credentials.NewTLS(&tls.Config{
		//	ServerName:   "server.razeen.me",
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	})

	dialOpt := grpc.WithTransportCredentials(transportCreds)

	// change to actual domain name
	conn, err := grpc.Dial("test.com:8080", dialOpt)
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
