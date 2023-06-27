package grpcutil

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/cpusoft/goutil/belogs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// grpcServer: a.com:8080
// should defer conn.Close()
func InitGrpcTcpClient(grpcServer string) (*grpc.ClientConn, error) {
	belogs.Debug("InitGrpcTcpClient(): grpcServer:", grpcServer)
	conn, err := grpc.Dial(grpcServer, grpc.WithInsecure())
	if err != nil {
		belogs.Error("StartRpcTcpClient():Dial fail:", err)
		return nil, err
	}
	return conn, nil
}

// grpcServer: a.com:8080
// defer conn.Close()
func InitGrpcTlsClient(grpcServer string,
	rootCrtFileName, publicCrtFileName, privateKeyFileName string) (*grpc.ClientConn, error) {
	belogs.Debug("InitGrpcTlsClient(): grpcServer:", grpcServer, "  rootCrtFileName:", rootCrtFileName,
		"  publicCrtFileName:", publicCrtFileName, "  privateKeyFileName:", privateKeyFileName)

	cert, err := tls.LoadX509KeyPair(publicCrtFileName,
		privateKeyFileName)
	if err != nil {
		belogs.Error("InitGrpcTlsClient(): LoadX509KeyPair fail:", err)
		return nil, err
	}

	certPool := x509.NewCertPool()
	bs, err := ioutil.ReadFile(rootCrtFileName)
	if err != nil {
		belogs.Error("InitGrpcTlsClient(): ReadFile fail:", err)
		return nil, err
	}

	if !certPool.AppendCertsFromPEM(bs) {
		belogs.Error("InitGrpcTlsClient(): AppendCertsFromPEM fail:")
		return nil, errors.New("client append cert from pem fail")
	}

	transportCreds := credentials.NewTLS(&tls.Config{
		//	ServerName:   "server.razeen.me",
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	})

	dialOpt := grpc.WithTransportCredentials(transportCreds)

	// change to actual domain name
	conn, err := grpc.Dial(grpcServer, dialOpt)
	if err != nil {
		fmt.Println("InitGrpcTlsClient(): Dial fail:", err)
		return nil, err
	}
	return conn, nil
}
