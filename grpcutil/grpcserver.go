package grpcutil

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"

	"github.com/cpusoft/goutil/belogs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// port:":8080"
func InitGrpcTcpServer(port string) (net.Listener, *grpc.Server, error) {
	belogs.Debug("InitGrpcTcpServer(): port:", port)
	listener, err := net.Listen("tcp", port)
	if err != nil {
		belogs.Error("InitGrpcTcpServer(): Listen fail, port:", port, err)
		return nil, nil, err
	}
	grpcServer := grpc.NewServer()
	return listener, grpcServer, nil
}

func StartGrpcTcpServer(listener net.Listener, grpcServer *grpc.Server) error {
	reflection.Register(grpcServer)
	belogs.Debug("StartGrpcTcpServer():start, listener:", listener, "  grpcServer:", grpcServer)
	err := grpcServer.Serve(listener)
	if err != nil {
		belogs.Error("StartGrpcTcpServer(): Serve fail,", err)
		return err
	}
	return nil
}

// port:":8080"
func InitGrpcTlsServer(port string,
	rootCrtFileName, publicCrtFileName, privateKeyFileName string) (net.Listener, *grpc.Server, error) {
	belogs.Debug("InitGrpcTlsServer(): port:", port, "  rootCrtFileName:", rootCrtFileName,
		"  publicCrtFileName:", publicCrtFileName, "  privateKeyFileName:", privateKeyFileName)
	listener, err := net.Listen("tcp", port)
	if err != nil {
		belogs.Error("InitGrpcTlsServer(): Listen fail, port:", port, err)
		return nil, nil, err
	}

	cert, err := tls.LoadX509KeyPair(publicCrtFileName, privateKeyFileName)
	if err != nil {
		belogs.Error("InitGrpcTlsServer(): LoadX509KeyPair fail, ",
			"  publicCrtFileName:", publicCrtFileName, "  privateKeyFileName:", privateKeyFileName, err)
		return nil, nil, err
	}

	certPool := x509.NewCertPool()
	rootBuf, err := ioutil.ReadFile(rootCrtFileName)
	if err != nil {
		belogs.Error("InitGrpcTlsServer(): ReadFile fail, rootCrtFileName:", rootCrtFileName, err)
		return nil, nil, err
	}

	if !certPool.AppendCertsFromPEM(rootBuf) {
		belogs.Error("InitGrpcTlsServer(): AppendCertsFromPEM fail,", err)
		return nil, nil, errors.New("server append certs from root fail")
	}

	tlsConf := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    certPool,
	}

	serverOpt := grpc.Creds(credentials.NewTLS(tlsConf))
	grpcServer := grpc.NewServer(serverOpt)
	return listener, grpcServer, nil
}

func StartGrpcTlsServer(listener net.Listener, grpcServer *grpc.Server) error {
	err := grpcServer.Serve(listener)
	belogs.Debug("StartGrpcTlsServer():start, listener:", listener, "  grpcServer:", grpcServer)
	if err != nil {
		belogs.Error("StartGrpcTlsServer(): Listen fail,", err)
		return err
	}
	return nil
}
