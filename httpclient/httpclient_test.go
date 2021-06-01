package httpclient

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestGetHttp(t *testing.T) {
	/*
		//往rp发送请求
		resp, body, err := GetHttp("http://localhost:8080/hello")

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("resp:", resp)
		fmt.Println("body:", body)
	*/
}

func TestGetHttps(t *testing.T) {
	/*
		//往rp发送请求
		resp, body, err := GetHttps("https://localhost:8081/hello")

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("resp:", resp)
		fmt.Println("body:", body)
	*/
}

// only in linux
func TestGetHttpsRrdp(t *testing.T) {
	/*
		//往rp发送请求
		url := `https://rrdp.arin.net/8fe05c2e-047d-49e7-8398-cd4250a572b1/18593/snapshot.xml`
		resutl, err := GetByCurl(url)

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("resp:", resutl)
	*/
}

func TestPostFile(t *testing.T) {
	/*
		resp, body, err := PostFile("http", "localhost", 8080, "/parse/start",
			`G:/Download/cert/cache/trustanchors/rpki.apnic.net/repository/apnic-rpki-root-iana-origin.cer`, `file`)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("resp:", resp)
		fmt.Println("body:", body)
	*/
}

func TestHttpsPostFile(t *testing.T) {
	/*
		resp, body, err := PostFileHttp("https://202.173.14.104:8071/slurm/upload",
			`G:\Download\rpstir2-std.log`, `file`)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("resp:", resp)
		fmt.Println("body:", body)
	*/
}

func TestHttpUnMarshalJson(t *testing.T) {

	url := `http://202.173.14.103:58085/allReset`
	var httpResponse HttpResponse
	err := PostAndUnmarshalStruct(url, "", false, &httpResponse)
	fmt.Println(httpResponse, err)

	resp, body, err := Post(url, "", false)
	resp.Body.Close()
	var httpResponse1 HttpResponse
	err = jsonutil.UnmarshalJson(body, &httpResponse1)
	fmt.Println(httpResponse1, err)

}

type HttpResponse struct {
	Result string      `json:"result"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}
