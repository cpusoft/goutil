package tlsserver

import (
	"bytes"
	"crypto/tls"
	"errors"
	"testing"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/tcpserverclient/util"
	"github.com/onsi/gomega/gstruct/errors"
)

type ServerProcessFunc struct {
}

func (spf *ServerProcessFunc) OnConnectProcess(tlsConn *tls.Conn) (err error) {
	peerCerts := tlsConn.ConnectionState().PeerCertificates
	if peerCerts == nil || len(peerCerts) == 0 {
		return errors.New("perrCerts is emtpy")
	}
	// The first element is the leaf certificate that the connection is verified against
	clientCert := peerCerts[0]

	subject := clientCert.Subject.CommonName
	belogs.Debug("OnConnectProcess(): spf: subject:", subject)
	dnsNames := clientCert.DNSNames
	belogs.Info("OnConnectProcess(): spf: dnsNames:", dnsNames)

	// can active send msg to client
	return nil
}
func (spf *ServerProcessFunc) ReceiveAndSendProcess(tlsConn *tls.Conn, receiveData []byte) (nextConnectPolicy int, leftData []byte, err error) {
	belogs.Debug("ReceiveAndSendProcess(): tlsserver  len(receiveData):", len(receiveData), "   receiveData:", convert.Bytes2String(receiveData))
	// need recombine
	packets, leftData, err := util.RecombineReceiveData(receiveData, PDU_TYPE_MIN_LEN, PDU_TYPE_LENGTH_START, PDU_TYPE_LENGTH_END)
	if err != nil {
		belogs.Error("ReceiveAndSendProcess(): tlsserver  RecombineReceiveData fail:", err)
		return util.NEXT_CONNECT_POLICE_CLOSE_FORCIBLE, nil, err
	}
	belogs.Debug("ReceiveAndSendProcess(): tlsserver  RecombineReceiveData packets.Len():", packets.Len())

	if packets == nil || packets.Len() == 0 {
		belogs.Debug("ReceiveAndSendProcess(): tlsserver  RecombineReceiveData packets is empty:  len(leftData):", len(leftData))
		return util.NEXT_CONNECT_POLICE_CLOSE_GRACEFUL, leftData, nil
	}
	sendData := make([]byte, 0)
	for e := packets.Front(); e != nil; e = e.Next() {
		packet, ok := e.Value.([]byte)
		if !ok || packet == nil || len(packet) == 0 {
			belogs.Debug("ReceiveAndSendProcess(): tlsserver  for packets fail:", convert.ToString(e.Value))
			break
		}
		tmpData, err := RtrProcess(packet)
		if err != nil {
			belogs.Error("ReceiveAndSendProcess(): tlsserver  RtrProcess fail:", err)
			return util.NEXT_CONNECT_POLICE_CLOSE_FORCIBLE, nil, err
		}
		sendData = append(sendData, tmpData...)

	}

	// may send or not send
	_, err = tlsConn.Write(sendData)
	if err != nil {
		belogs.Error("ReceiveAndSendProcess(): tlsserver ,  Write fail:  tcpConn:", tcpConn.RemoteAddr().String(), err)
		return util.NEXT_CONNECT_POLICE_CLOSE_FORCIBLE, nil, err
	}
	// continue to receive next receiveData
	return util.NEXT_CONNECT_POLICE_KEEP, leftData, nil
}
func (spf *ServerProcessFunc) OnCloseProcess(tlsConn *tls.Conn) {

}
func (spf *ServerProcessFunc) ActiveSendProcess(tlsConn *tls.Conn, sendData []byte) (err error) {
	return
}

func TestCreateTlsServer(t *testing.T) {
	serverProcessFunc := new(ServerProcessFunc)
	rootCrtFileName := `./ca/ca.cer`
	publicCrtFileName := `./server/server.cer`
	privateKeyFileName := `./server/serverkey.pem`

	ts := NewTlsServer(rootCrtFileName, publicCrtFileName, privateKeyFileName, true, serverProcessFunc)
	ts.Start("9999")
}
func GetTlsServerData() (buffer []byte) {

	return []byte{0x00, 0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
}

func RtrProcess(receiveData []byte) (sendData []byte, err error) {
	buf := bytes.NewReader(receiveData)
	belogs.Debug("RtrProcess(): tlsserver buf:", buf)
	return GetTlsServerData(), nil
}

const (
	PDU_TYPE_MIN_LEN      = 8
	PDU_TYPE_LENGTH_START = 4
	PDU_TYPE_LENGTH_END   = 8
)
