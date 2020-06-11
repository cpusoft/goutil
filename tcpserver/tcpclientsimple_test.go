package tcpserver

import (
	"fmt"
	"testing"
)

func TestClientSendAndReceive(t *testing.T) {
	server := "202.173.14.105:8082"
	sendData := "serialNotify"
	receiveData, err := ClientSendAndReceive(server, []byte(sendData))
	fmt.Println(receiveData, err)
}
