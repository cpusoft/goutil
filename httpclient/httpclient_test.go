package httpclient

import (
	"fmt"
	"testing"

	httpclient "github.com/cpusoft/goutil/httpclient"
)

func TestGetHttp(t *testing.T) {
	//往rp发送请求
	resp, body, err := httpclient.GetHttp("http://localhost:8080/hello")

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("resp:", resp)
	fmt.Println("body:", body)
}

func TestGetHttps(t *testing.T) {
	//往rp发送请求
	resp, body, err := httpclient.GetHttps("https://localhost:8081/hello")

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("resp:", resp)
	fmt.Println("body:", body)
}

func TestPostFile(t *testing.T) {
	resp, body, err := httpclient.PostFile("http", "localhost", 8080, "/parsecert/upload",
		`apnic-rpki-root-iana-origin.cer`, `G:\Download\cert\cache\trustanchors\rpki.apnic.net\repository\apnic-rpki-root-iana-origin.cer`)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("resp:", resp)
	fmt.Println("body:", body)
}
