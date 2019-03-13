package httpclient

import (
	"fmt"
	"testing"

	httpclient "github.com/cpusoft/goutil/httpclient"
)

func TestGet(t *testing.T) {
	//往rp发送请求
	resp, body, err := httpclient.Get("http://localhost:8080/hello")

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("resp: %+v", resp)
	fmt.Println("body: %+v", body)
}

func TestGetTLS(t *testing.T) {
	//往rp发送请求
	resp, body, err := httpclient.GetTLS("https://localhost:8081/hello")

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("resp: %+v", resp)
	fmt.Println("body: %+v", body)
}
