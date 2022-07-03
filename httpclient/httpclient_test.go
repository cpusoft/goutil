package httpclient

import (
	"fmt"
	"testing"
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

	//往rp发送请求
	url := `https://rrdp.afrinic.net/notification.xml` //https://rpki.august.tw/rrdp/notification.xml`
	resutl, err := GetByCurl(url)

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("resp:", resutl)

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

	resp, body, err := PostFileHttp("http://202.173.14.105:8070/parsevalidate/parsefilesimple",
		`G:\Download\cert\oWhEB7GUTj5ZqlXo7X2VbNrJ9xw.cer`, `file`)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("resp:", resp)
	fmt.Println("body:", body)

}

func TestHttpUnMarshalJson(t *testing.T) {
	/*
		url := `http://202.173.14.103:58085/allReset`
		var httpResponse HttpResponse
		err := PostAndUnmarshalStruct(url, "", false, &httpResponse)
		fmt.Println(httpResponse, err)

		resp, body, err := Post(url, "", false)
		resp.Body.Close()
		var httpResponse1 HttpResponse
		err = jsonutil.UnmarshalJson(body, &httpResponse1)
		fmt.Println(httpResponse1, err)
	*/
}

type HttpResponse struct {
	Result string      `json:"result"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}
